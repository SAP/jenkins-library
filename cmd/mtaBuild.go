package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// for mocking
var downloadAndCopySettingsFiles = maven.DownloadAndCopySettingsFiles

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
		return -1, fmt.Errorf("Unknown BuildTarget/Platform: '%s'", str)
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

func mtaBuild(config mtaBuildOptions,
	telemetryData *telemetry.CustomData,
	commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment) {
	log.Entry().Debugf("Launching mta build")
	files := piperutils.Files{}
	httpClient := piperhttp.Client{}
	e := command.Command{}

	npmExecutorOptions := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry, ExecRunner: &e}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)

	err := runMtaBuild(config, commonPipelineEnvironment, &e, &files, &httpClient, npmExecutor)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute mta build")
	}
}

func runMtaBuild(config mtaBuildOptions,
	commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment,
	e command.ExecRunner,
	p piperutils.FileUtils,
	httpClient piperhttp.Downloader,
	npmExecutor npm.Executor) error {

	e.Stdout(log.Writer()) // not sure if using the logging framework here is a suitable approach. We handover already log formatted
	e.Stderr(log.Writer()) // entries to a logging framework again. But this is considered to be some kind of project standard.

	var err error

	err = handleSettingsFiles(config, p, httpClient)
	if err != nil {
		return err
	}

	err = npmExecutor.SetNpmRegistries()

	mtaYamlFile := "mta.yaml"
	mtaYamlFileExists, err := p.FileExists(mtaYamlFile)

	if err != nil {
		return err
	}

	if !mtaYamlFileExists {

		if err = createMtaYamlFile(mtaYamlFile, config.ApplicationName, p); err != nil {
			return err
		}

	} else {
		log.Entry().Infof("\"%s\" file found in project sources", mtaYamlFile)
	}

	if err = setTimeStamp(mtaYamlFile, p); err != nil {
		return err
	}

	mtarName, err := getMtarName(config, mtaYamlFile, p)

	if err != nil {
		return err
	}

	var call []string

	switch config.MtaBuildTool {

	case "classic":

		mtaJar := getMarJarName(config)

		buildTarget, err := ValueOfBuildTarget(config.BuildTarget)

		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return err
		}

		call = append(call, "java", "-jar", mtaJar, "--mtar", mtarName, fmt.Sprintf("--build-target=%s", buildTarget), "build")
		if len(config.Extensions) != 0 {
			call = append(call, fmt.Sprintf("--extension=%s", config.Extensions))
		}

	case "cloudMbt":

		platform, err := ValueOfBuildTarget(config.Platform)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return err
		}

		call = append(call, "mbt", "build", "--mtar", mtarName, "--platform", platform.String())
		if len(config.Extensions) != 0 {
			call = append(call, fmt.Sprintf("--extensions=%s", config.Extensions))
		}
		call = append(call, "--target", "./")

	default:

		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("Unknown mta build tool: \"%s\"", config.MtaBuildTool)
	}

	if err = addNpmBinToPath(e); err != nil {
		return err
	}

	if len(config.M2Path) > 0 {
		absolutePath, err := p.Abs(config.M2Path)
		if err != nil {
			return err
		}
		e.AppendEnv([]string{"MAVEN_OPTS=-Dmaven.repo.local=" + absolutePath})
	}

	log.Entry().Infof("Executing mta build call: \"%s\"", strings.Join(call, " "))

	if err := e.RunExecutable(call[0], call[1:]...); err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return err
	}

	commonPipelineEnvironment.mtarFilePath = mtarName

	if config.InstallArtifacts {
		// install maven artifacts in local maven repo because `mbt build` executes `mvn package -B`
		err = installMavenArtifacts(e, config)
		if err != nil {
			return err
		}
		// mta-builder executes 'npm install --production', therefore we need 'npm ci/install' to install the dev-dependencies
		err = npmExecutor.InstallAllDependencies(npmExecutor.FindPackageJSONFiles())
		if err != nil {
			return err
		}
	}
	return err
}

func installMavenArtifacts(e command.ExecRunner, config mtaBuildOptions) error {
	pomXMLExists, err := piperutils.FileExists("pom.xml")
	if err != nil {
		return err
	}
	if pomXMLExists {
		err = maven.InstallMavenArtifacts(e, maven.EvaluateOptions{M2Path: config.M2Path})
		if err != nil {
			return err
		}
	}
	return nil
}

func getMarJarName(config mtaBuildOptions) string {

	mtaJar := "mta.jar"

	if len(config.MtaJarLocation) > 0 {
		mtaJar = config.MtaJarLocation
	}

	return mtaJar
}

func addNpmBinToPath(e command.ExecRunner) error {
	dir, _ := os.Getwd()
	newPath := path.Join(dir, "node_modules", ".bin")
	oldPath := os.Getenv("PATH")
	if len(oldPath) > 0 {
		newPath = newPath + ":" + oldPath
	}
	e.SetEnv([]string{"PATH=" + newPath})
	return nil
}

func getMtarName(config mtaBuildOptions, mtaYamlFile string, p piperutils.FileUtils) (string, error) {

	mtarName := config.MtarName
	if len(mtarName) == 0 {

		log.Entry().Debugf("mtar name not provided via config. Extracting from file \"%s\"", mtaYamlFile)

		mtaID, err := getMtaID(mtaYamlFile, p)

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

func setTimeStamp(mtaYamlFile string, p piperutils.FileUtils) error {

	mtaYaml, err := p.FileRead(mtaYamlFile)
	if err != nil {
		return err
	}

	mtaYamlStr := string(mtaYaml)

	timestampVar := "${timestamp}"
	if strings.Contains(mtaYamlStr, timestampVar) {

		if err := p.FileWrite(mtaYamlFile, []byte(strings.ReplaceAll(mtaYamlStr, timestampVar, getTimestamp())), 0644); err != nil {
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

func createMtaYamlFile(mtaYamlFile, applicationName string, p piperutils.FileUtils) error {

	log.Entry().Debugf("mta yaml file not found in project sources.")

	if len(applicationName) == 0 {
		return fmt.Errorf("'%[1]s' not found in project sources and 'applicationName' not provided as parameter - cannot generate '%[1]s' file", mtaYamlFile)
	}

	packageFileExists, err := p.FileExists("package.json")
	if !packageFileExists {
		return fmt.Errorf("package.json file does not exist")
	}

	var result map[string]interface{}
	pContent, err := p.FileRead("package.json")
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

	p.FileWrite(mtaYamlFile, []byte(mtaConfig), 0644)
	log.Entry().Infof("\"%s\" created.", mtaYamlFile)

	return nil
}

func handleSettingsFiles(config mtaBuildOptions,
	p piperutils.FileUtils,
	httpClient piperhttp.Downloader) error {

	return downloadAndCopySettingsFiles(config.GlobalSettingsFile, config.ProjectSettingsFile, p, httpClient)
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

func getMtaID(mtaYamlFile string, fileUtils piperutils.FileUtils) (string, error) {

	var result map[string]interface{}
	p, err := fileUtils.FileRead(mtaYamlFile)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(p, &result)
	if err != nil {
		return "", err
	}

	id, ok := result["ID"].(string)
	if !ok || len(id) == 0 {
		fmt.Errorf("Id not found in mta yaml file (or wrong type)")
	}

	return id, nil
}
