package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/yaml"
	"github.com/elliotchance/orderedmap"
	"github.com/pkg/errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type cfFileUtil interface {
	FileExists(string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	Getwd() (string, error)
	Glob(string) ([]string, error)
}

var _now = time.Now
var _cfLogin = cfLogin
var _cfLogout = cfLogout
var _getManifest = getManifest
var _substitute = yaml.Substitute
var fileUtils cfFileUtil = piperutils.Files{}

// for simplify mocking. Maybe we find a more elegant way (mock for CFUtils)
func cfLogin(c command.ExecRunner, options cloudfoundry.LoginOptions) error {
	cf := &cloudfoundry.CFUtils{Exec: c}
	return cf.Login(options)
}

// for simplify mocking. Maybe we find a more elegant way (mock for CFUtils)
func cfLogout(c command.ExecRunner) error {
	cf := &cloudfoundry.CFUtils{Exec: c}
	return cf.Logout()
}

const defaultSmokeTestScript = `#!/usr/bin/env bash
# this is simply testing if the application root returns HTTP STATUS_CODE
curl -so /dev/null -w '%{response_code}' https://$1 | grep $STATUS_CODE`

func cloudFoundryDeploy(config cloudFoundryDeployOptions, telemetryData *telemetry.CustomData, influxData *cloudFoundryDeployInflux) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runCloudFoundryDeploy(&config, telemetryData, influxData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %s", err)
	}
}

func runCloudFoundryDeploy(config *cloudFoundryDeployOptions, telemetryData *telemetry.CustomData, influxData *cloudFoundryDeployInflux, command command.ExecRunner) error {

	log.Entry().Infof("General parameters: deployTool='%s', deployType='%s', cfApiEndpoint='%s', cfOrg='%s', cfSpace='%s'",
		config.DeployTool, config.DeployType, config.APIEndpoint, config.Org, config.Space)

	var deployTriggered bool
	var err error

	if config.DeployTool == "mtaDeployPlugin" {
		deployTriggered = true
		err = handleMTADeployment(config, command)
	} else if config.DeployTool == "cf_native" {
		deployTriggered = true
		err = handleCFNativeDeployment(config, command)
	} else {
		log.Entry().Warningf("Found unsupported deployTool ('%s'). Skipping deployment. Supported deploy tools: 'mtaDeployPlugin', 'cf_native'", config.DeployTool)
	}

	if deployTriggered {
		prepareInflux(err == nil, config, influxData)
	}

	return err
}

func prepareInflux(success bool, config *cloudFoundryDeployOptions, influxData *cloudFoundryDeployInflux) error {

	if influxData == nil {
		return nil
	}

	result := "FAILURE"

	if success {
		result = "SUCCESS"
	}

	// TODO done via CPE in groovy. In case we need this we have to enhance CPE
	influxData.deployment_data.tags.artifactVersion = "<n/a>"
	influxData.deployment_data.tags.deployUser = config.Username
	influxData.deployment_data.tags.deployResult = result
	influxData.deployment_data.tags.cfAPIEndpoint = config.APIEndpoint
	influxData.deployment_data.tags.cfOrg = config.Org
	influxData.deployment_data.tags.cfSpace = config.Space

	// n/a (literally) is also reported in groovy
	influxData.deployment_data.fields.artifactURL = "n/a"

	influxData.deployment_data.fields.deployTime = strings.ToUpper(_now().Format("Jan 02 2006 15:04:05"))

	// we should discuss how we handle the job trigger
	// 1.) outside Jenkins
	// 2.) inside Jenkins (how to get)
	influxData.deployment_data.fields.jobTrigger = "<n/a>"

	return nil
}

func handleMTADeployment(config *cloudFoundryDeployOptions, command command.ExecRunner) error {

	mtarFilePath := config.MtaPath

	if len(mtarFilePath) == 0 {

		var err error
		mtarFilePath, err = findMtar()

		if err != nil {
			return err
		}

		log.Entry().Debugf("Using mtar file '%s' found in workspace", mtarFilePath)

	} else {

		exists, err := fileUtils.FileExists(mtarFilePath)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("mtar file '%s' retrieved from configuration does not exist", mtarFilePath)
		}

		log.Entry().Debugf("Using mtar file '%s' from configuration", mtarFilePath)
	}

	return deployMta(config, mtarFilePath, command)
}

type deployConfig struct {
	DeployCommand   string
	DeployOptions   []string
	AppName         string
	ManifestFile    string
	SmokeTestScript []string
}

func handleCFNativeDeployment(config *cloudFoundryDeployOptions, command command.ExecRunner) error {

	deployType, err := checkAndUpdateDeployTypeForNotSupportedManifest(config)

	if err != nil {
		return err
	}

	var appName string
	var deployCommand string
	var smokeTestScript []string
	var deployOptions []string

	// deploy command will be provided by the prepare functions below

	if deployType == "blue-green" {
		deployCommand, deployOptions, smokeTestScript, err = prepareBlueGreenCfNativeDeploy(config)
		if err != nil {
			return errors.Wrapf(err, "Cannot prepare cf native deployment. DeployType '%s'", deployType)
		}
	} else if deployType == "standard" {
		deployCommand, deployOptions, smokeTestScript, err = prepareCfPushCfNativeDeploy(config)
		if err != nil {
			return errors.Wrapf(err, "Cannot prepare cf push native deployment. DeployType '%s'", deployType)
		}
	} else {
		return fmt.Errorf("Invalid deploy type received: '%s'. Supported values: %v", deployType, []string{"blue-green", "standard"})
	}

	if len(config.AppName) == 0 {
		// Basically we try to retrieve the app name from the manifest here since it is not provided from the config
		// Later on we don't use the app name retrieved here since we can use it from the manifest.
		// Here we simply fail early when the app name is not provided and also not contained in the manifest.
		appName, err = getAppNameOrFail(config)
		if err != nil {
			return err
		}
	}

	log.Entry().Infof("CF native deployment ('%s') with:", config.DeployType)
	log.Entry().Infof("cfAppName='%s'", appName)
	log.Entry().Infof("cfManifest='%s'", config.Manifest)
	log.Entry().Infof("cfManifestVariables: '%v'", config.ManifestVariables)
	log.Entry().Infof("cfManifestVariablesFiles: '%v'", config.ManifestVariablesFiles)
	log.Entry().Infof("cfdeployDockerImage: '%s'", config.DeployDockerImage)
	log.Entry().Infof("smokeTestScript: '%s'", config.SmokeTestScript)

	additionalEnvironment := []string{
		"STATUS_CODE=" + strconv.FormatInt(int64(config.SmokeTestStatusCode), 10),
	}

	if len(config.DockerPassword) > 0 {
		additionalEnvironment = append(additionalEnvironment, "CF_DOCKER_PASSWORD="+config.DockerPassword)
	}

	myDeployConfig := deployConfig{
		DeployCommand:   deployCommand,
		DeployOptions:   deployOptions,
		AppName:         config.AppName,
		ManifestFile:    config.Manifest,
		SmokeTestScript: smokeTestScript,
	}

	log.Entry().Infof("DeployConfig: %v", myDeployConfig)

	//return nil
	return deployCfNative(myDeployConfig, config, additionalEnvironment, command)
}

func deployCfNative(deployConfig deployConfig, config *cloudFoundryDeployOptions, additionalEnvironment []string, cmd command.ExecRunner) error {

	deployStatement := []string{
		deployConfig.DeployCommand,
	}

	if len(deployConfig.AppName) > 0 {
		deployStatement = append(deployStatement, deployConfig.AppName)
	}

	if len(deployConfig.DeployOptions) > 0 {
		deployStatement = append(deployStatement, deployConfig.DeployOptions...)
	}

	if len(deployConfig.ManifestFile) > 0 {
		deployStatement = append(deployStatement, "-f")
		deployStatement = append(deployStatement, deployConfig.ManifestFile)
	}

	if len(config.DeployDockerImage) > 0 && config.DeployType != "blue-green" {
		deployStatement = append(deployStatement, "--docker-image", config.DeployDockerImage)
	}

	if len(config.DockerUsername) > 0 && config.DeployType != "blue-green" {
		deployStatement = append(deployStatement, "--docker-username", config.DockerUsername)
	}

	if len(deployConfig.SmokeTestScript) > 0 {
		deployStatement = append(deployStatement, deployConfig.SmokeTestScript...)
	}

	if len(config.CfNativeDeployParameters) > 0 {
		deployStatement = append(deployStatement, strings.FieldsFunc(config.CfNativeDeployParameters, func(c rune) bool {
			return c == ' '
		})...)
	}

	stopOldAppIfRunning := func(_cmd command.ExecRunner) error {

		if config.KeepOldInstance && config.DeployType == "blue-green" {
			oldAppName := deployConfig.AppName + "-old"
			_ = oldAppName

			var buff bytes.Buffer

			oldOut := _cmd.Stdout
			_ = oldOut
			_cmd.Stdout(&buff)

			defer func() {
				_cmd.Stdout(log.Writer())
			}()

			err := _cmd.RunExecutable("cf", "stop", oldAppName)

			if err != nil {

				cfStopLog := buff.String()

				if !strings.Contains(cfStopLog, oldAppName+" not found") {
					return fmt.Errorf("Could not stop application %s. Error: %s", oldAppName, cfStopLog)
				}
				log.Entry().Infof("Cannot stop application '%s': %s", oldAppName, cfStopLog)

			} else {
				log.Entry().Infof("Old application '%s' has been stopped.", oldAppName)
			}
		}

		return nil
	}

	return cfDeploy(config, deployStatement, additionalEnvironment, stopOldAppIfRunning, cmd)
}

func getManifest(name string) (cloudfoundry.Manifest, error) {
	return cloudfoundry.ReadManifest(name)
}

func getAppNameOrFail(config *cloudFoundryDeployOptions) (string, error) {

	if len(config.AppName) > 0 {
		return config.AppName, nil
	}
	if config.DeployType == "blue-green" {
		return "", fmt.Errorf("Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)")
	}
	if len(config.Manifest) == 0 {
		return "", fmt.Errorf("Manifest file not provided in configuration. Cannot retrieve app name")
	}
	fileExists, err := fileUtils.FileExists(config.Manifest)
	if err != nil {
		return "", err
	}
	if !fileExists {
		return "", fmt.Errorf("Manifest file '%s' not found. Cannot retrieve app name", config.Manifest)
	}
	m, err := _getManifest(config.Manifest)
	if err != nil {
		return "", err
	}
	apps, err := m.GetApplications()
	if err != nil {
		return "", err
	}
	if len(apps) > 0 {
		namePropertyExists, err := m.ApplicationHasProperty(0, "name")
		if err != nil {
			return "", err
		}
		if namePropertyExists {
			appName, err := m.GetApplicationProperty(0, "name")
			if err != nil {
				return "", err
			}
			if name, ok := appName.(string); ok {
				if len(name) > 0 {
					return name, nil
				}
			}
		}
	}

	return "", fmt.Errorf("No appName available in manifest '%s'", config.Manifest)
}

func handleSmokeTestScript(smokeTestScript string) ([]string, error) {

	if smokeTestScript == "blueGreenCheckScript.sh" {
		// what should we do if there is already a script with the given name? Should we really overwrite ...
		err := fileUtils.FileWrite(smokeTestScript, []byte(defaultSmokeTestScript), 0755)
		if err != nil {
			return []string{}, err
		}
		log.Entry().Debugf("smoke test script '%s' has been written.", smokeTestScript)
	}

	if len(smokeTestScript) > 0 {
		err := os.Chmod(smokeTestScript, 0755)
		if err != nil {
			return []string{}, err
		}
		pwd, err := fileUtils.Getwd()
		if err != nil {
			return []string{}, err
		}

		return []string{"--smoke-test", fmt.Sprintf("%s/%s", pwd, smokeTestScript)}, nil
	}
	return []string{}, nil
}

func prepareBlueGreenCfNativeDeploy(config *cloudFoundryDeployOptions) (string, []string, []string, error) {

	smokeTest, err := handleSmokeTestScript(config.SmokeTestScript)
	if err != nil {
		return "", []string{}, []string{}, err
	}

	var deployOptions = []string{}

	if !config.KeepOldInstance {
		deployOptions = append(deployOptions, "--delete-old-apps")
	}

	if len(config.Manifest) > 0 {
		manifestFileExists, err := fileUtils.FileExists(config.Manifest)
		if err != nil {
			return "", []string{}, []string{}, err
		}

		if !manifestFileExists {

			log.Entry().Infof("Manifest file '%s' does not exist", config.Manifest)

		} else {

			manifestVariables, err := toStringInterfaceMap(toParameterMap(config.ManifestVariables))
			if err != nil {
				return "", []string{}, []string{}, err
			}

			modified, err := _substitute(config.Manifest, manifestVariables, config.ManifestVariablesFiles)
			if err != nil {
				return "", []string{}, []string{}, errors.Wrap(err, "Cannot prepare manifest file")
			}

			if modified {
				log.Entry().Infof("Manifest file '%s' has been updated (variable substitution)", config.Manifest)
			} else {
				log.Entry().Infof("Manifest file '%s' has not been updated (variable substitution)", config.Manifest)
			}

			err = handleLegacyCfManifest(config.Manifest)
			if err != nil {
				return "", []string{}, []string{}, err
			}
		}
	} else {
		log.Entry().Info("No manifest file configured")
	}
	return "blue-green-deploy", deployOptions, smokeTest, nil
}

func toParameterMap(parameters []string) (*orderedmap.OrderedMap, error) {

	parameterMap := orderedmap.NewOrderedMap()

	for _, p := range parameters {
		keyVal := strings.Split(p, "=")
		if len(keyVal) != 2 {
			return nil, fmt.Errorf("Invalid parameter provided (expected format <key>=<val>: '%s'", p)
		}
		parameterMap.Set(keyVal[0], keyVal[1])
	}
	return parameterMap, nil
}

func handleLegacyCfManifest(manifestFile string) error {
	manifest, err := _getManifest(manifestFile)
	if err != nil {
		return err
	}

	err = manifest.Transform()
	if err != nil {
		return err
	}
	if manifest.IsModified() {

		err = manifest.WriteManifest()

		if err != nil {
			return err
		}
		log.Entry().Infof("Manifest file '%s' was in legacy format has been transformed and updated.", manifestFile)
	} else {
		log.Entry().Infof("Manifest file '%s' was not in legacy format. No tranformation needed, no update performed.", manifestFile)
	}
	return nil
}

func prepareCfPushCfNativeDeploy(config *cloudFoundryDeployOptions) (string, []string, []string, error) {

	deployOptions := []string{}
	varOptions, err := getVarOptions(config.ManifestVariables)
	if err != nil {
		return "", []string{}, []string{}, err
	}
	varFileOptions, err := getVarFileOptions(config.ManifestVariablesFiles)
	if err != nil {
		return "", []string{}, []string{}, err
	}

	deployOptions = append(deployOptions, varOptions...)
	deployOptions = append(deployOptions, varFileOptions...)

	return "push", deployOptions, []string{}, nil
}

func toStringInterfaceMap(in *orderedmap.OrderedMap, err error) (map[string]interface{}, error) {

	out := map[string]interface{}{}

	if err == nil {
		for _, key := range in.Keys() {
			if k, ok := key.(string); ok {
				val, exists := in.Get(key)
				if exists {
					out[k] = val
				} else {
					return map[string]interface{}{}, fmt.Errorf("No entry found for '%v'", key)
				}
			} else {
				return map[string]interface{}{}, fmt.Errorf("Cannot cast key '%v' to string", key)
			}
		}
	}

	return out, err
}
func getVarOptions(vars []string) ([]string, error) {

	varsMap, err := toParameterMap(vars)
	if err != nil {
		return []string{}, err
	}

	varsResult := []string{}

	for _, key := range varsMap.Keys() {
		val, _ := varsMap.Get(key)
		if v, ok := val.(string); ok {
			varsResult = append(varsResult, "--var", fmt.Sprintf("%s=%s", key, quoteAndBashEscape(v)))
		} else {
			return []string{}, fmt.Errorf("Cannot cast '%v' to string", val)
		}
	}
	return varsResult, nil
}

func getVarFileOptions(varFiles []string) ([]string, error) {

	varFilesResult := []string{}

	for _, varFile := range varFiles {
		fExists, err := fileUtils.FileExists(varFile)
		if err != nil {
			return []string{}, err
		}

		if !fExists {
			log.Entry().Warningf("We skip adding not-existing file '%s' as a vars-file to the cf create-service-push call", varFile)
			continue
		}

		varFilesResult = append(varFilesResult, "--vars-file", quoteAndBashEscape(varFile))
	}

	if len(varFilesResult) > 0 {
		log.Entry().Infof("We will add the following string to the cf push call: '%s'", strings.Join(varFilesResult, " "))
	}
	return varFilesResult, nil
}

func quoteAndBashEscape(s string) string {
	escapedSingleQuote := "'\"'\"'"
	return strings.ReplaceAll(s, "'", escapedSingleQuote)
}

func checkAndUpdateDeployTypeForNotSupportedManifest(config *cloudFoundryDeployOptions) (string, error) {

	manifestFile := config.Manifest
	if len(manifestFile) == 0 {
		manifestFile = "manifest.yml"
	}
	var manifestFileExists bool
	var err error

	if len(manifestFile) > 0 {
		manifestFileExists, err = fileUtils.FileExists(manifestFile)
		if err != nil {
			return "", err
		}
	}

	if config.DeployType == "blue-green" && manifestFileExists {

		m, _ := _getManifest(manifestFile)

		apps, err := m.GetApplications()

		if err != nil {
			return "", err
		}
		if len(apps) > 1 {
			return "", fmt.Errorf("Your manifest contains more than one application. For blue green deployments your manifest file may contain only one application")
		}

		hasNoRouteProperty, err := m.ApplicationHasProperty(0, "no-route")
		if err != nil {
			return "", err
		}
		if len(apps) == 1 && hasNoRouteProperty {

			const deployTypeStandard = "standard"
			log.Entry().Warningf("Blue green deployment is not possible for application without route. Using deployment type '%s' instead.", deployTypeStandard)
			return deployTypeStandard, nil
		}
	}

	return config.DeployType, nil
}

func deployMta(config *cloudFoundryDeployOptions, mtarFilePath string, command command.ExecRunner) error {

	deployCommand := "deploy"
	deployParams := []string{}

	if len(config.MtaDeployParameters) > 0 {
		deployParams = append(deployParams, strings.Split(config.MtaDeployParameters, " ")...)
	}

	if config.DeployType == "bg-deploy" || config.DeployType == "blue-green" {

		deployCommand = "bg-deploy"

		const noConfirmFlag = "--no-confirm"
		if !contains(deployParams, noConfirmFlag) {
			deployParams = append(deployParams, noConfirmFlag)
		}
	}

	cfDeployParams := []string{
		deployCommand,
		mtarFilePath,
	}

	if len(deployParams) > 0 {
		cfDeployParams = append(cfDeployParams, deployParams...)
	}

	{
		var mtaExtensionDescriptor string

		if len(config.MtaExtensionDescriptor) > 0 {
			if !strings.HasPrefix(config.MtaExtensionDescriptor, "-e") {
				mtaExtensionDescriptor = "-e " + config.MtaExtensionDescriptor
			} else {
				mtaExtensionDescriptor = config.MtaExtensionDescriptor
			}
		}

		if len(mtaExtensionDescriptor) > 0 {
			cfDeployParams = append(cfDeployParams, mtaExtensionDescriptor)
		}
	}

	return cfDeploy(config, cfDeployParams, nil, nil, command)
}

// would make sense to have that method in some kind of helper instead having it here
func contains(collection []string, key string) bool {

	for _, v := range collection {
		if v == key {
			return true
		}
	}
	return false
}

func cfDeploy(
	config *cloudFoundryDeployOptions,
	cfDeployParams []string,
	additionalEnvironment []string,
	postDeployAction func(command command.ExecRunner) error,
	command command.ExecRunner) error {

	const cfLogFile = "cf.log"
	var err error
	var loginPerformed bool

	if additionalEnvironment == nil {
		additionalEnvironment = []string{}
	}
	additionalEnvironment = append(additionalEnvironment, "CF_TRACE="+cfLogFile)

	if len(config.CfHome) > 0 {
		additionalEnvironment = append(additionalEnvironment, "CF_HOME="+config.CfHome)
	}

	if len(config.CfPluginHome) > 0 {
		additionalEnvironment = append(additionalEnvironment, "CF_PLUGIN_HOME="+config.CfPluginHome)
	}

	log.Entry().Infof("Using additional environment variables: %s", additionalEnvironment)

	// TODO set HOME to config.DockerWorkspace
	command.SetEnv(additionalEnvironment)

	err = _cfLogin(command, cloudfoundry.LoginOptions{
		CfAPIEndpoint: config.APIEndpoint,
		CfOrg:         config.Org,
		CfSpace:       config.Space,
		Username:      config.Username,
		Password:      config.Password,
		CfAPIOpts:     splitAtWhitespace(config.APIParameters),
		CfLoginOpts:   splitAtWhitespace(config.LoginParameters),
	})

	if err == nil {
		loginPerformed = true
	}

	if err == nil {
		err = command.RunExecutable("cf", []string{"plugins"}...)
		if err != nil {
			log.Entry().WithError(err).Errorf("Command '%s' failed.", []string{"plugins"})
		}
	}

	if err == nil {
		err = command.RunExecutable("cf", cfDeployParams...)
		if err != nil {
			log.Entry().WithError(err).Errorf("Command '%s' failed.", cfDeployParams)
		}
	}

	if err == nil && postDeployAction != nil {
		err = postDeployAction(command)
	}

	if loginPerformed {

		logoutErr := _cfLogout(command)

		if logoutErr != nil {
			log.Entry().WithError(logoutErr).Errorf("Cannot perform cf logout")
			if err == nil {
				err = logoutErr
			}
		}
	}

	if err != nil || GeneralConfig.Verbose {
		e := handleCfCliLog(cfLogFile)
		if e != nil {
			log.Entry().WithError(err).Errorf("Error reading cf log file '%s'.", cfLogFile)
		}
	}

	return err
}

func findMtar() (string, error) {

	const pattern = "**/*.mtar"

	mtars, err := fileUtils.Glob(pattern)

	if err != nil {
		return "", err
	}

	if len(mtars) == 0 {
		return "", fmt.Errorf("No mtar file matching pattern '%s' found", pattern)
	}

	if len(mtars) > 1 {
		sMtars := []string{}
		for _, mtar := range mtars {
			sMtars = append(sMtars, mtar)
		}
		return "", fmt.Errorf("Found multiple mtar files matching pattern '%s' (%s), please specify file via mtaPath parameter 'mtarPath'", pattern, strings.Join(sMtars, ","))
	}

	return mtars[0], nil
}

func handleCfCliLog(logFile string) error {

	fExists, err := fileUtils.FileExists(logFile)

	if err != nil {
		return err
	}

	log.Entry().Info("### START OF CF CLI TRACE OUTPUT ###")

	if fExists {

		f, err := os.Open(logFile)

		if err != nil {
			return err
		}

		bReader := bufio.NewReader(f)

		var done bool

		for {
			line, err := bReader.ReadString('\n')
			if err == io.EOF {

				// maybe inappropriate to log as info. Maybe the line from the
				// log indicates an error, but that is something like a project
				// standard.
				done = true
			} else if err != nil {

				break
			}

			log.Entry().Info(strings.TrimSuffix(line, "\n"))

			if done {
				break
			}
		}

	} else {
		log.Entry().Warningf("No trace file found at '%s'", logFile)
	}

	log.Entry().Info("### END OF CF CLI TRACE OUTPUT ###")

	return err
}

func splitAtWhitespace(s string) []string {
	if len(strings.TrimSpace(s)) == 0 {
		return []string{}
	}
	return regexp.MustCompile("\\s+").Split(s, -1)
}
