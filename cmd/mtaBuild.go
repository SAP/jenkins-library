package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/SAP/jenkins-library/pkg/npm"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const templateMtaYml = `_schema-version: "3.1"
ID: "{{.ID}}"
version: {{.Version}}

parameters:
  hcp-deployer-version: "1.1.0"

modules:
  - name: {{.ApplicationName}}
    type: html5
    path: .
    parameters:
       version: {{.Version}}-${timestamp}
       name: {{.ApplicationName}}
    build-parameters:
      builder: grunt
      build-result: dist`

// MTABuildTarget ...
type MTABuildTarget int

const (
	// NEO ...
	NEO MTABuildTarget = iota
	// CF ...
	CF MTABuildTarget = iota
	//XSA ...
	XSA MTABuildTarget = iota
)

// ValueOfBuildTarget ...
func ValueOfBuildTarget(str string) (MTABuildTarget, error) {
	switch str {
	case "NEO":
		return NEO, nil
	case "CF":
		return CF, nil
	case "XSA":
		return XSA, nil
	default:
		return -1, fmt.Errorf("Unknown Platform: '%s'", str)
	}
}

// String ...
func (m MTABuildTarget) String() string {
	return [...]string{
		"NEO",
		"CF",
		"XSA",
	}[m]
}

type mtaBuildUtils interface {
	maven.Utils

	SetEnv(env []string)
	AppendEnv(env []string)

	Abs(path string) (string, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error

	DownloadAndCopySettingsFiles(globalSettingsFile string, projectSettingsFile string) error

	SetNpmRegistries(defaultNpmRegistry string) error
	InstallAllDependencies(defaultNpmRegistry string) error
}

type mtaBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func (bundle *mtaBuildUtilsBundle) SetNpmRegistries(defaultNpmRegistry string) error {
	npmExecutorOptions := npm.ExecutorOptions{DefaultNpmRegistry: defaultNpmRegistry, ExecRunner: bundle}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)
	return npmExecutor.SetNpmRegistries()
}

func (bundle *mtaBuildUtilsBundle) InstallAllDependencies(defaultNpmRegistry string) error {
	npmExecutorOptions := npm.ExecutorOptions{DefaultNpmRegistry: defaultNpmRegistry, ExecRunner: bundle}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)
	return npmExecutor.InstallAllDependencies(npmExecutor.FindPackageJSONFiles())
}

func (bundle *mtaBuildUtilsBundle) DownloadAndCopySettingsFiles(globalSettingsFile string, projectSettingsFile string) error {
	return maven.DownloadAndCopySettingsFiles(globalSettingsFile, projectSettingsFile, bundle)
}

func newMtaBuildUtilsBundle() mtaBuildUtils {
	utils := mtaBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func mtaBuild(config mtaBuildOptions,
	telemetryData *telemetry.CustomData,
	commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment) {
	log.Entry().Debugf("Launching mta build")
	utils := newMtaBuildUtilsBundle()

	err := runMtaBuild(config, commonPipelineEnvironment, utils)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute mta build")
	}
}

func runMtaBuild(config mtaBuildOptions,
	commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment,
	utils mtaBuildUtils) error {

	var err error

	err = handleSettingsFiles(config, utils)
	if err != nil {
		return err
	}

	err = handleActiveProfileUpdate(config, utils)
	if err != nil {
		return err
	}

	err = utils.SetNpmRegistries(config.DefaultNpmRegistry)

	mtaYamlFile := "mta.yaml"
	mtaYamlFileExists, err := utils.FileExists(mtaYamlFile)

	if err != nil {
		return err
	}

	if !mtaYamlFileExists {

		if err = createMtaYamlFile(mtaYamlFile, config.ApplicationName, utils); err != nil {
			return err
		}

	} else {
		log.Entry().Infof("\"%s\" file found in project sources", mtaYamlFile)
	}

	if err = setTimeStamp(mtaYamlFile, utils); err != nil {
		return err
	}

	mtarName, err := getMtarName(config, mtaYamlFile, utils)

	if err != nil {
		return err
	}

	var call []string

	platform, err := ValueOfBuildTarget(config.Platform)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return err
	}

	call = append(call, "mbt", "build", "--mtar", mtarName, "--platform", platform.String())
	if len(config.Extensions) != 0 {
		call = append(call, fmt.Sprintf("--extensions=%s", config.Extensions))
	}
	if config.Source != "" && config.Source != "./" {
		call = append(call, "--source", config.Source)
	} else {
		call = append(call, "--source", "./")
	}
	if config.Target != "" && config.Target != "./" {
		call = append(call, "--target", config.Target)
	} else {
		call = append(call, "--target", "./")
	}

	if err = addNpmBinToPath(utils); err != nil {
		return err
	}

	if len(config.M2Path) > 0 {
		absolutePath, err := utils.Abs(config.M2Path)
		if err != nil {
			return err
		}
		utils.AppendEnv([]string{"MAVEN_OPTS=-Dmaven.repo.local=" + absolutePath})
	}

	log.Entry().Infof("Executing mta build call: \"%s\"", strings.Join(call, " "))

	if err := utils.RunExecutable(call[0], call[1:]...); err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return err
	}

	commonPipelineEnvironment.mtarFilePath = mtarName

	if config.InstallArtifacts {
		// install maven artifacts in local maven repo because `mbt build` executes `mvn package -B`
		err = installMavenArtifacts(utils, config)
		if err != nil {
			return err
		}
		// mta-builder executes 'npm install --production', therefore we need 'npm ci/install' to install the dev-dependencies
		err = utils.InstallAllDependencies(config.DefaultNpmRegistry)
		if err != nil {
			return err
		}
	}

	if config.Publish {
		log.Entry().Infof("publish detected")
		if (len(config.MtaDeploymentRepositoryPassword) > 0) && (len(config.MtaDeploymentRepositoryUser) > 0) &&
			(len(config.MtaDeploymentRepositoryURL) > 0) {
			if (len(config.MtarGroup) > 0) && (len(config.Version) > 0) {
				downloadClient := &piperhttp.Client{}

				credentialsEncoded := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", config.MtaDeploymentRepositoryUser, config.MtaDeploymentRepositoryPassword)))
				headers := http.Header{}
				headers.Add("Authorization", credentialsEncoded)

				config.MtarGroup = strings.ReplaceAll(config.MtarGroup, ".", "/")

				mtarArtifactName := mtarName

				mtarArtifactName = strings.ReplaceAll(mtarArtifactName, ".mtar", "")
				mtarArtifactName = strings.ReplaceAll(mtarArtifactName, ".", "/")

				config.MtaDeploymentRepositoryURL += config.MtarGroup + "/" + mtarArtifactName + "/" + config.Version + "/" + fmt.Sprintf("%v-%v.%v", mtarArtifactName, config.Version, "mtar")

				commonPipelineEnvironment.custom.mtarPublishedURL = config.MtaDeploymentRepositoryURL

				log.Entry().Infof("pushing mtar artifact to repository : %s", config.MtaDeploymentRepositoryURL)

				_, httpErr := downloadClient.UploadRequest(http.MethodPut, config.MtaDeploymentRepositoryURL, mtarName, mtarName, headers, nil)
				if httpErr != nil {
					return errors.Wrap(err, "failed to upload mtar to repository")
				}
			} else {
				return errors.New("mtarGroup, version not found and must be present")

			}

		} else {
			return errors.New("altDeploymentRepositoryUser, altDeploymentRepositoryPassword and altDeploymentRepositoryURL not found , must be present")
		}
	} else {
		log.Entry().Infof("no publish detected, skipping upload of mtar artifact")
	}
	return err
}

func handleActiveProfileUpdate(config mtaBuildOptions, utils mtaBuildUtils) error {
	if len(config.Profiles) > 0 {
		return maven.UpdateActiveProfileInSettingsXML(config.Profiles, utils)
	}
	return nil
}

func installMavenArtifacts(utils mtaBuildUtils, config mtaBuildOptions) error {
	pomXMLExists, err := utils.FileExists("pom.xml")
	if err != nil {
		return err
	}
	if pomXMLExists {
		err = maven.InstallMavenArtifacts(&maven.EvaluateOptions{M2Path: config.M2Path}, utils)
		if err != nil {
			return err
		}
	}
	return nil
}

func addNpmBinToPath(utils mtaBuildUtils) error {
	dir, _ := os.Getwd()
	newPath := path.Join(dir, "node_modules", ".bin")
	oldPath := os.Getenv("PATH")
	if len(oldPath) > 0 {
		newPath = newPath + ":" + oldPath
	}
	utils.SetEnv([]string{"PATH=" + newPath})
	return nil
}

func getMtarName(config mtaBuildOptions, mtaYamlFile string, utils mtaBuildUtils) (string, error) {

	mtarName := config.MtarName
	if len(mtarName) == 0 {

		log.Entry().Debugf("mtar name not provided via config. Extracting from file \"%s\"", mtaYamlFile)

		mtaID, err := getMtaID(mtaYamlFile, utils)

		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", err
		}

		if len(mtaID) == 0 {
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", fmt.Errorf("Invalid mtar ID. Was empty")
		}

		log.Entry().Debugf("mtar name extracted from file \"%s\": \"%s\"", mtaYamlFile, mtaID)

		mtarName = mtaID + ".mtar"
	}

	return mtarName, nil

}

func setTimeStamp(mtaYamlFile string, utils mtaBuildUtils) error {

	mtaYaml, err := utils.FileRead(mtaYamlFile)
	if err != nil {
		return err
	}

	mtaYamlStr := string(mtaYaml)

	timestampVar := "${timestamp}"
	if strings.Contains(mtaYamlStr, timestampVar) {

		if err := utils.FileWrite(mtaYamlFile, []byte(strings.ReplaceAll(mtaYamlStr, timestampVar, getTimestamp())), 0644); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return err
		}
		log.Entry().Infof("Timestamp replaced in \"%s\"", mtaYamlFile)
	} else {
		log.Entry().Infof("No timestamp contained in \"%s\". File has not been modified.", mtaYamlFile)
	}

	return nil
}

func getTimestamp() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func createMtaYamlFile(mtaYamlFile, applicationName string, utils mtaBuildUtils) error {

	log.Entry().Debugf("mta yaml file not found in project sources.")

	if len(applicationName) == 0 {
		return fmt.Errorf("'%[1]s' not found in project sources and 'applicationName' not provided as parameter - cannot generate '%[1]s' file", mtaYamlFile)
	}

	packageFileExists, err := utils.FileExists("package.json")
	if !packageFileExists {
		return fmt.Errorf("package.json file does not exist")
	}

	var result map[string]interface{}
	pContent, err := utils.FileRead("package.json")
	if err != nil {
		return err
	}
	json.Unmarshal(pContent, &result)

	version, ok := result["version"].(string)
	if !ok {
		return fmt.Errorf("Version not found in \"package.json\" (or wrong type)")
	}

	name, ok := result["name"].(string)
	if !ok {
		return fmt.Errorf("Name not found in \"package.json\" (or wrong type)")
	}

	mtaConfig, err := generateMta(name, applicationName, version)
	if err != nil {
		return err
	}

	utils.FileWrite(mtaYamlFile, []byte(mtaConfig), 0644)
	log.Entry().Infof("\"%s\" created.", mtaYamlFile)

	return nil
}

func handleSettingsFiles(config mtaBuildOptions, utils mtaBuildUtils) error {
	return utils.DownloadAndCopySettingsFiles(config.GlobalSettingsFile, config.ProjectSettingsFile)
}

func generateMta(id, applicationName, version string) (string, error) {

	if len(id) == 0 {
		return "", fmt.Errorf("Generating mta file: ID not provided")
	}
	if len(applicationName) == 0 {
		return "", fmt.Errorf("Generating mta file: ApplicationName not provided")
	}
	if len(version) == 0 {
		return "", fmt.Errorf("Generating mta file: Version not provided")
	}

	tmpl, e := template.New("mta.yaml").Parse(templateMtaYml)
	if e != nil {
		return "", e
	}

	type properties struct {
		ID              string
		ApplicationName string
		Version         string
	}

	props := properties{ID: id, ApplicationName: applicationName, Version: version}

	var script bytes.Buffer
	tmpl.Execute(&script, props)
	return script.String(), nil
}

func getMtaID(mtaYamlFile string, utils mtaBuildUtils) (string, error) {

	var result map[string]interface{}
	p, err := utils.FileRead(mtaYamlFile)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(p, &result)
	if err != nil {
		return "", err
	}

	id, ok := result["ID"].(string)
	if !ok || len(id) == 0 {
		return "", fmt.Errorf("Id not found in mta yaml file (or wrong type)")
	}

	return id, nil
}
