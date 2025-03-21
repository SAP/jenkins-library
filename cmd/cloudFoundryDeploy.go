package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/yaml"
	"github.com/pkg/errors"
)

type cfFileUtil interface {
	FileExists(string) (bool, error)
	FileRename(string, string) error
	FileRead(string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	Getwd() (string, error)
	Glob(string) ([]string, error)
	Chmod(string, os.FileMode) error
	Copy(string, string) (int64, error)
	Stat(path string) (os.FileInfo, error)
}

var _now = time.Now
var _cfLogin = cfLogin
var _cfLogout = cfLogout
var _getManifest = getManifest
var _replaceVariables = yaml.Substitute
var _getVarsOptions = cloudfoundry.GetVarsOptions
var _getVarsFileOptions = cloudfoundry.GetVarsFileOptions
var _environ = os.Environ
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

	err := validateAppName(config.AppName)

	if err != nil {
		return err
	}

	validateDeployTool(config)

	var deployTriggered bool

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

func validateDeployTool(config *cloudFoundryDeployOptions) {
	if config.DeployTool != "" || config.BuildTool == "" {
		return
	}

	switch config.BuildTool {
	case "mta":
		config.DeployTool = "mtaDeployPlugin"
	default:
		config.DeployTool = "cf_native"
	}
	log.Entry().Infof("Parameter deployTool not specified - deriving from buildTool '%s': '%s'",
		config.BuildTool, config.DeployTool)
}

func validateAppName(appName string) error {
	// for the sake of brevity we consider the empty string as valid app name here
	isValidAppName, err := regexp.MatchString("^$|^[a-zA-Z0-9]$|^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$", appName)
	if err != nil {
		return err
	}
	if isValidAppName {
		return nil
	}
	const (
		underscore = "_"
		dash       = "-"
		docuLink   = "https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings"
	)

	log.Entry().Warningf("Your application name '%s' contains non-alphanumeric characters which may lead to errors in the future, "+
		"as they are not supported by CloudFoundry. For more details please visit %s", appName, docuLink)

	var fail bool
	message := []string{fmt.Sprintf("Your application name '%s'", appName)}
	if strings.Contains(appName, underscore) {
		message = append(message, fmt.Sprintf("contains a '%s' (underscore) which is not allowed, only letters, dashes and numbers can be used.", underscore))
		fail = true
	}
	if strings.HasPrefix(appName, dash) || strings.HasSuffix(appName, dash) {
		message = append(message, fmt.Sprintf("starts or ends with a '%s' (dash) which is not allowed, only letters and numbers can be used.", dash))
		fail = true
	}
	message = append(message, fmt.Sprintf("Please change the name to fit this requirement(s). For more details please visit %s.", docuLink))
	if fail {
		return errors.New(strings.Join(message, " "))
	}
	return nil
}

func prepareInflux(success bool, config *cloudFoundryDeployOptions, influxData *cloudFoundryDeployInflux) {

	if influxData == nil {
		return
	}

	result := "FAILURE"

	if success {
		result = "SUCCESS"
	}

	influxData.deployment_data.tags.artifactVersion = config.ArtifactVersion
	influxData.deployment_data.tags.deployUser = config.Username
	influxData.deployment_data.tags.deployResult = result
	influxData.deployment_data.tags.cfAPIEndpoint = config.APIEndpoint
	influxData.deployment_data.tags.cfOrg = config.Org
	influxData.deployment_data.tags.cfSpace = config.Space

	// n/a (literally) is also reported in groovy
	influxData.deployment_data.fields.artifactURL = "n/a"
	influxData.deployment_data.fields.commitHash = config.CommitHash

	influxData.deployment_data.fields.deployTime = strings.ToUpper(_now().Format("Jan 02 2006 15:04:05"))

	// we should discuss how we handle the job trigger
	// 1.) outside Jenkins
	// 2.) inside Jenkins (how to get)
	influxData.deployment_data.fields.jobTrigger = "n/a"
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
			return errors.Wrapf(err, "Cannot check if file path '%s' exists", mtarFilePath)
		}

		if !exists {
			return fmt.Errorf("mtar file '%s' retrieved from configuration does not exist", mtarFilePath)
		}

		log.Entry().Debugf("Using mtar file '%s' from configuration", mtarFilePath)
	}

	return deployMta(config, mtarFilePath, command)
}

type deployConfig struct {
	DeployCommand string
	DeployOptions []string
	AppName       string
	ManifestFile  string
}

func handleCFNativeDeployment(config *cloudFoundryDeployOptions, command command.ExecRunner) error {

	var deployCommand string
	var deployOptions []string
	var err error

	// deploy command will be provided by the prepare functions below
	if config.DeployType == "blue-green" {
		return fmt.Errorf("Blue-green deployment type is deprecated for cf native builds." +
			"Instead set parameter `cfNativeDeployParameters: '--strategy rolling'`. " +
			"Please refer to the Cloud Foundry documentation for further information: " +
			"https://docs.cloudfoundry.org/devguide/deploy-apps/rolling-deploy.html." +
			"Or alternatively, switch to mta build tool. Please refer to mta build tool" +
			"documentation for further information: https://sap.github.io/cloud-mta-build-tool/configuration/.")
	} else if config.DeployType == "standard" {
		deployCommand, deployOptions, err = prepareCfPushCfNativeDeploy(config)
		if err != nil {
			return errors.Wrapf(err, "Cannot prepare cf push native deployment. DeployType '%s'", config.DeployType)
		}
	} else {
		return fmt.Errorf("Invalid deploy type received: '%s'. Supported value: standard", config.DeployType)
	}

	appName, err := getAppName(config)
	if err != nil {
		return err
	}

	manifestFile, err := getManifestFileName(config)

	log.Entry().Infof("CF native deployment ('%s') with:", config.DeployType)
	log.Entry().Infof("cfAppName='%s'", appName)
	log.Entry().Infof("cfManifest='%s'", manifestFile)
	log.Entry().Infof("cfManifestVariables: '%v'", config.ManifestVariables)
	log.Entry().Infof("cfManifestVariablesFiles: '%v'", config.ManifestVariablesFiles)
	log.Entry().Infof("cfdeployDockerImage: '%s'", config.DeployDockerImage)

	var additionalEnvironment []string

	if len(config.DockerPassword) > 0 {
		additionalEnvironment = []string{("CF_DOCKER_PASSWORD=" + config.DockerPassword)}
	}

	myDeployConfig := deployConfig{
		DeployCommand: deployCommand,
		DeployOptions: deployOptions,
		AppName:       config.AppName,
		ManifestFile:  config.Manifest,
	}

	log.Entry().Infof("DeployConfig: %v", myDeployConfig)

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

	if len(config.DeployDockerImage) > 0 {
		deployStatement = append(deployStatement, "--docker-image", config.DeployDockerImage)
	}

	if len(config.DockerUsername) > 0 {
		deployStatement = append(deployStatement, "--docker-username", config.DockerUsername)
	}

	if len(config.CfNativeDeployParameters) > 0 {
		deployStatement = append(deployStatement, strings.Fields(config.CfNativeDeployParameters)...)
	}

	return cfDeploy(config, deployStatement, additionalEnvironment, cmd)
}

func getManifest(name string) (cloudfoundry.Manifest, error) {
	return cloudfoundry.ReadManifest(name)
}

func getManifestFileName(config *cloudFoundryDeployOptions) (string, error) {

	manifestFileName := config.Manifest
	if len(manifestFileName) == 0 {
		manifestFileName = "manifest.yml"
	}
	return manifestFileName, nil
}

func getAppName(config *cloudFoundryDeployOptions) (string, error) {

	if len(config.AppName) > 0 {
		return config.AppName, nil
	}

	manifestFile, err := getManifestFileName(config)

	fileExists, err := fileUtils.FileExists(manifestFile)
	if err != nil {
		return "", errors.Wrapf(err, "Cannot check if file '%s' exists", manifestFile)
	}
	if !fileExists {
		return "", fmt.Errorf("Manifest file '%s' not found. Cannot retrieve app name", manifestFile)
	}
	manifest, err := _getManifest(manifestFile)
	if err != nil {
		return "", err
	}
	apps, err := manifest.GetApplications()
	if err != nil {
		return "", err
	}

	if len(apps) == 0 {
		return "", fmt.Errorf("No apps declared in manifest '%s'", manifestFile)
	}
	namePropertyExists, err := manifest.ApplicationHasProperty(0, "name")
	if err != nil {
		return "", err
	}
	if !namePropertyExists {
		return "", fmt.Errorf("No appName available in manifest '%s'", manifestFile)
	}
	appName, err := manifest.GetApplicationProperty(0, "name")
	if err != nil {
		return "", err
	}
	var name string
	var ok bool
	if name, ok = appName.(string); !ok {
		return "", fmt.Errorf("appName from manifest '%s' has wrong type", manifestFile)
	}
	if len(name) == 0 {
		return "", fmt.Errorf("appName from manifest '%s' is empty", manifestFile)
	}
	return name, nil
}

func prepareCfPushCfNativeDeploy(config *cloudFoundryDeployOptions) (string, []string, error) {

	deployOptions := []string{}
	varOptions, err := _getVarsOptions(config.ManifestVariables)
	if err != nil {
		return "", []string{}, errors.Wrapf(err, "Cannot prepare var-options: '%v'", config.ManifestVariables)
	}

	varFileOptions, err := _getVarsFileOptions(config.ManifestVariablesFiles)
	if err != nil {
		if e, ok := err.(*cloudfoundry.VarsFilesNotFoundError); ok {
			for _, missingVarFile := range e.MissingFiles {
				log.Entry().Warningf("We skip adding not-existing file '%s' as a vars-file to the cf create-service-push call", missingVarFile)
			}
		} else {
			return "", []string{}, errors.Wrapf(err, "Cannot prepare var-file-options: '%v'", config.ManifestVariablesFiles)
		}
	}

	deployOptions = append(deployOptions, varOptions...)
	deployOptions = append(deployOptions, varFileOptions...)

	return "push", deployOptions, nil
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
		if !slices.Contains(deployParams, noConfirmFlag) {
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

	extFileParams, extFiles := handleMtaExtensionDescriptors(config.MtaExtensionDescriptor)

	for _, extFile := range extFiles {
		_, err := fileUtils.Copy(extFile, extFile+".original")
		if err != nil {
			return fmt.Errorf("Cannot prepare mta extension files: %w", err)
		}
		_, _, err = handleMtaExtensionCredentials(extFile, config.MtaExtensionCredentials)
		if err != nil {
			return fmt.Errorf("Cannot handle credentials inside mta extension files: %w", err)
		}
	}

	cfDeployParams = append(cfDeployParams, extFileParams...)

	err := cfDeploy(config, cfDeployParams, nil, command)

	for _, extFile := range extFiles {
		renameError := fileUtils.FileRename(extFile+".original", extFile)
		if err == nil && renameError != nil {
			return renameError
		}
	}

	return err
}

func handleMtaExtensionCredentials(extFile string, credentials map[string]interface{}) (updated, containsUnresolved bool, err error) {

	log.Entry().Debugf("Inserting credentials into extension file '%s'", extFile)

	b, err := fileUtils.FileRead(extFile)
	if err != nil {
		return false, false, errors.Wrapf(err, "Cannot handle credentials for mta extension file '%s'", extFile)
	}
	content := string(b)

	env, err := toMap(_environ(), "=")
	if err != nil {
		return false, false, errors.Wrap(err, "Cannot handle mta extension credentials.")
	}

	missingCredentials := []string{}
	for name, credentialKey := range credentials {
		credKey, ok := credentialKey.(string)
		if !ok {
			return false, false, fmt.Errorf("cannot handle mta extension credentials: Cannot cast '%v' (type %T) to string", credentialKey, credentialKey)
		}

		const allowedVariableNamePattern = "^[-_A-Za-z0-9]+$"
		alphaNumOnly := regexp.MustCompile(allowedVariableNamePattern)
		if !alphaNumOnly.MatchString(name) {
			return false, false, fmt.Errorf("credential key name '%s' contains unsupported character. Must contain only %s", name, allowedVariableNamePattern)
		}
		pattern := regexp.MustCompile("<%=\\s*" + name + "\\s*%>")
		if pattern.MatchString(content) {
			cred := env[toEnvVarKey(credKey)]
			if len(cred) == 0 {
				missingCredentials = append(missingCredentials, credKey)
				continue
			}
			content = pattern.ReplaceAllLiteralString(content, cred)
			updated = true
			log.Entry().Debugf("Mta extension credentials handling: Placeholder '%s' has been replaced by credential denoted by '%s'/'%s' in file '%s'", name, credKey, toEnvVarKey(credKey), extFile)
		} else {
			log.Entry().Debugf("Mta extension credentials handling: Variable '%s' is not used in file '%s'", name, extFile)
		}
	}
	if len(missingCredentials) > 0 {
		missinCredsEnvVarKeyCompatible := []string{}
		for _, missingKey := range missingCredentials {
			missinCredsEnvVarKeyCompatible = append(missinCredsEnvVarKeyCompatible, toEnvVarKey(missingKey))
		}
		// ensure stable order of the entries. Needed e.g. for the tests.
		sort.Strings(missingCredentials)
		sort.Strings(missinCredsEnvVarKeyCompatible)
		return false, false, fmt.Errorf("cannot handle mta extension credentials: No credentials found for '%s'/'%s'. Are these credentials maintained?", missingCredentials, missinCredsEnvVarKeyCompatible)
	}
	if !updated {
		log.Entry().Debugf("Mta extension credentials handling: Extension file '%s' has not been updated. Seems to contain no credentials.", extFile)
	} else {
		fInfo, err := fileUtils.Stat(extFile)
		fMode := fInfo.Mode()
		if err != nil {
			return false, false, errors.Wrap(err, "Cannot handle mta extension credentials.")
		}
		err = fileUtils.FileWrite(extFile, []byte(content), fMode)
		if err != nil {
			return false, false, errors.Wrap(err, "Cannot handle mta extension credentials.")
		}
		log.Entry().Debugf("Mta extension credentials handling: Extension file '%s' has been updated.", extFile)
	}

	re := regexp.MustCompile(`<%=.+%>`)
	placeholders := re.FindAll([]byte(content), -1)
	containsUnresolved = (len(placeholders) > 0)

	if containsUnresolved {
		log.Entry().Warningf("mta extension credential handling: Unresolved placeholders found after inserting credentials: %s", placeholders)
	}

	return updated, containsUnresolved, nil
}

func toEnvVarKey(key string) string {
	key = regexp.MustCompile(`[^A-Za-z0-9]`).ReplaceAllString(key, "_")
	return strings.ToUpper(regexp.MustCompile(`([a-z0-9])([A-Z])`).ReplaceAllString(key, "${1}_${2}"))
}

func toMap(keyValue []string, separator string) (map[string]string, error) {
	result := map[string]string{}
	for _, entry := range keyValue {
		kv := strings.Split(entry, separator)
		if len(kv) < 2 {
			return map[string]string{}, fmt.Errorf("Cannot convert to map: separator '%s' not found in entry '%s'", separator, entry)
		}
		result[kv[0]] = strings.Join(kv[1:], separator)
	}
	return result, nil
}

func handleMtaExtensionDescriptors(mtaExtensionDescriptor string) ([]string, []string) {
	var result = []string{}
	var extFiles = []string{}
	for _, part := range strings.Fields(strings.Trim(mtaExtensionDescriptor, " ")) {
		if part == "-e" || part == "" {
			continue
		}
		// REVISIT: maybe check if the extension descriptor exists
		extFiles = append(extFiles, part)
	}
	if len(extFiles) > 0 {
		result = append(result, "-e")
		result = append(result, strings.Join(extFiles, ","))
	}
	return result, extFiles
}

func cfDeploy(
	config *cloudFoundryDeployOptions,
	cfDeployParams []string,
	additionalEnvironment []string,
	command command.ExecRunner) error {

	const cfLogFile = "cf.log"
	var err error
	var loginPerformed bool

	// TODO: remove after testing?
	log.Entry().Infof("additionalEnvironment '%s'", additionalEnvironment)
	log.Entry().Infof("config.CfTrace '%s'", config.CfTrace)

	if config.CfTrace {
		additionalEnvironment = append(additionalEnvironment, "CF_TRACE="+cfLogFile)
	} else {
		additionalEnvironment = append(additionalEnvironment, "CF_TRACE=true") // Print API request diagnostics to stdout
	}

	if len(config.CfHome) > 0 {
		additionalEnvironment = append(additionalEnvironment, "CF_HOME="+config.CfHome)
	}

	if len(config.CfPluginHome) > 0 {
		additionalEnvironment = append(additionalEnvironment, "CF_PLUGIN_HOME="+config.CfPluginHome)
	}

	log.Entry().Infof("Using additional environment variables: %s", additionalEnvironment)

	// TODO set HOME to config.DockerWorkspace
	command.SetEnv(additionalEnvironment)

	err = command.RunExecutable("cf", "version")

	if err == nil {
		err = _cfLogin(command, cloudfoundry.LoginOptions{
			CfAPIEndpoint: config.APIEndpoint,
			CfOrg:         config.Org,
			CfSpace:       config.Space,
			Username:      config.Username,
			Password:      config.Password,
			CfLoginOpts:   strings.Fields(config.LoginParameters),
		})
	}

	if err == nil {
		loginPerformed = true
		err = command.RunExecutable("cf", []string{"plugins"}...)
		if err != nil {
			log.Entry().WithError(err).Errorf("Command '%s' failed.", []string{"plugins"})
		}
	}

	// TODO: remove after testing?
	fileExists, _ := piperutils.FileExists(cfLogFile);
	log.Entry().Infof("cfLogFile fileExists? '%s'", fileExists)

	if fileExists && !config.CfTrace {
		log.Entry().Infof("Removing existing cf log file '%s'", cfLogFile)
		err = os.Remove(cfLogFile)
		if err != nil {
			log.Entry().WithError(err).Errorf("Cannot remove existing cf log file '%s'", cfLogFile)
		}
	}

	if err == nil {
		err = command.RunExecutable("cf", cfDeployParams...)
		if err != nil {
			log.Entry().WithError(err).Errorf("Command '%s' failed.", cfDeployParams)
		}
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
		if config.CfTrace {
			if e := handleCfCliLog(cfLogFile); e != nil {
				log.Entry().WithError(err).Errorf("Error reading cf log file '%s': %v", cfLogFile, e)
			}
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
		sMtars = append(sMtars, mtars...)
		return "", fmt.Errorf("Found multiple mtar files matching pattern '%s' (%s), please specify file via parameter 'mtarPath'", pattern, strings.Join(sMtars, ","))
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

		defer f.Close()

		bReader := bufio.NewReader(f)
		for {
			line, err := bReader.ReadString('\n')
			if err == nil || err == io.EOF {
				// maybe inappropriate to log as info. Maybe the line from the
				// log indicates an error, but that is something like a project
				// standard.
				log.Entry().Info(strings.TrimSuffix(line, "\n"))
			}
			if err != nil {
				break
			}
		}
	} else {
		log.Entry().Warningf("No trace file found at '%s'", logFile)
	}

	log.Entry().Info("### END OF CF CLI TRACE OUTPUT ###")

	return err
}
