package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
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
    type: com.sap.hcp.html5
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
		Command: &command.Command{
			StepName: "mtaBuild",
		},
		Files:  &piperutils.Files{},
		Client: &piperhttp.Client{},
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

	mtaYamlFile := filepath.Join(getSourcePath(config), "mta.yaml")
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

	mtarName, isMtarNativelySuffixed, err := getMtarName(config, mtaYamlFile, utils)

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

	call = append(call, "--source", getSourcePath(config))
	call = append(call, "--target", getAbsPath(getMtarFileRoot(config)))

	if config.Jobs > 0 {
		call = append(call, "--mode=verbose")
		call = append(call, "--jobs="+strconv.Itoa(config.Jobs))
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

	log.Entry().Debugf("creating build settings information...")
	stepName := "mtaBuild"
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return err
	}

	mtaConfig := buildsettings.BuildOptions{
		Profiles:           config.Profiles,
		GlobalSettingsFile: config.GlobalSettingsFile,
		Publish:            config.Publish,
		BuildSettingsInfo:  config.BuildSettingsInfo,
		DefaultNpmRegistry: config.DefaultNpmRegistry,
		DockerImage:        dockerImage,
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&mtaConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

	commonPipelineEnvironment.mtarFilePath = filepath.ToSlash(getMtarFilePath(config, mtarName))
	commonPipelineEnvironment.custom.mtaBuildToolDesc = filepath.ToSlash(mtaYamlFile)

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
				httpClient := &piperhttp.Client{}

				credentialsEncoded := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", config.MtaDeploymentRepositoryUser, config.MtaDeploymentRepositoryPassword)))
				headers := http.Header{}
				headers.Add("Authorization", credentialsEncoded)

				config.MtarGroup = strings.ReplaceAll(config.MtarGroup, ".", "/")

				mtarArtifactName := mtarName

				// only trim the .mtar suffix from the mtarName
				if !isMtarNativelySuffixed {
					mtarArtifactName = strings.TrimSuffix(mtarArtifactName, ".mtar")
				}

				config.MtaDeploymentRepositoryURL += config.MtarGroup + "/" + mtarArtifactName + "/" + config.Version + "/" + fmt.Sprintf("%v-%v.%v", mtarArtifactName, config.Version, "mtar")

				commonPipelineEnvironment.custom.mtarPublishedURL = config.MtaDeploymentRepositoryURL

				log.Entry().Infof("pushing mtar artifact to repository : %s", config.MtaDeploymentRepositoryURL)

				data, err := os.Open(getMtarFilePath(config, mtarName))
				if err != nil {
					return errors.Wrap(err, "failed to open mtar archive for upload")
				}
				_, httpErr := httpClient.SendRequest("PUT", config.MtaDeploymentRepositoryURL, data, headers, nil)

				if httpErr != nil {
					return errors.Wrap(err, "failed to upload mtar to repository")
				}
			} else {
				return errors.New("mtarGroup, version not found and must be present")

			}

		} else {
			return errors.New("mtaDeploymentRepositoryUser, mtaDeploymentRepositoryPassword and mtaDeploymentRepositoryURL not found , must be present")
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

func getMtarName(config mtaBuildOptions, mtaYamlFile string, utils mtaBuildUtils) (string, bool, error) {

	mtarName := config.MtarName
	isMtarNativelySuffixed := false
	if len(mtarName) == 0 {

		log.Entry().Debugf("mtar name not provided via config. Extracting from file \"%s\"", mtaYamlFile)

		mtaID, err := getMtaID(mtaYamlFile, utils)

		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", isMtarNativelySuffixed, err
		}

		if len(mtaID) == 0 {
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", isMtarNativelySuffixed, fmt.Errorf("Invalid mtar ID. Was empty")
		}

		log.Entry().Debugf("mtar name extracted from file \"%s\": \"%s\"", mtaYamlFile, mtaID)

		// there can be cases where the mtaId itself has the value com.myComapany.mtar , adding an extra .mtar causes .mtar.mtar
		if !strings.HasSuffix(mtaID, ".mtar") {
			mtarName = mtaID + ".mtar"
		} else {
			isMtarNativelySuffixed = true
			mtarName = mtaID
		}

	}

	return mtarName, isMtarNativelySuffixed, nil

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

	log.Entry().Infof("\"%s\" file not found in project sources", mtaYamlFile)

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

	if err := utils.FileWrite(mtaYamlFile, []byte(mtaConfig), 0644); err != nil {
		return fmt.Errorf("failed to write %v: %w", mtaYamlFile, err)
	}
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
	if err := tmpl.Execute(&script, props); err != nil {
		log.Entry().Warningf("failed to execute template: %v", err)
	}
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

// the "source" path locates the project's root
func getSourcePath(config mtaBuildOptions) string {
	path := config.Source
	if path == "" {
		path = "./"
	}
	return filepath.FromSlash(path)
}

// target defines a subfolder of the project's root
func getTargetPath(config mtaBuildOptions) string {
	path := config.Target
	if path == "" {
		path = "./"
	}
	return filepath.FromSlash(path)
}

// the "mtar" path resides below the project's root
// path=<config.source>/<config.target>/<mtarname>
func getMtarFileRoot(config mtaBuildOptions) string {
	sourcePath := getSourcePath(config)
	targetPath := getTargetPath(config)

	return filepath.FromSlash(filepath.Join(sourcePath, targetPath))
}

func getMtarFilePath(config mtaBuildOptions, mtarName string) string {
	root := getMtarFileRoot(config)

	if root == "" || root == filepath.FromSlash("./") {
		return mtarName
	}

	return filepath.FromSlash(filepath.Join(root, mtarName))
}

func getAbsPath(path string) string {
	abspath, err := filepath.Abs(path)
	// ignore error, pass customers path value in case of trouble
	if err != nil {
		abspath = path
	}
	return filepath.FromSlash(abspath)
}
