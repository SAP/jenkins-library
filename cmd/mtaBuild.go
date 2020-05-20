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
var getSettingsFile = maven.GetSettingsFile

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
	err := runMtaBuild(config, commonPipelineEnvironment, &command.Command{}, &files, &httpClient)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute mta build")
	}
}

func runMtaBuild(config mtaBuildOptions,
	commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment,
	e execRunner,
	p piperutils.FileUtils,
	httpClient piperhttp.Downloader) error {

	e.Stdout(log.Writer()) // not sure if using the logging framework here is a suitable approach. We handover already log formatted
	e.Stderr(log.Writer()) // entries to a logging framework again. But this is considered to be some kind of project standard.

	var err error

	err = handleSettingsFiles(config, p, httpClient)
	if err != nil {
		return err
	}

	err = configureNpmRegistry(config.DefaultNpmRegistry, "default", "", e)
	if err != nil {
		return err
	}
	err = configureNpmRegistry(config.SapNpmRegistry, "SAP", "@sap", e)
	if err != nil {
		return err
	}

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
			return err
		}

		call = append(call, "java", "-jar", mtaJar, "--mtar", mtarName, fmt.Sprintf("--build-target=%s", buildTarget), "build")
		if len(config.Extensions) != 0 {
			call = append(call, fmt.Sprintf("--extension=%s", config.Extensions))
		}

	case "cloudMbt":

		platform, err := ValueOfBuildTarget(config.Platform)
		if err != nil {
			return err
		}

		call = append(call, "mbt", "build", "--mtar", mtarName, "--platform", platform.String())
		if len(config.Extensions) != 0 {
			call = append(call, fmt.Sprintf("--extensions=%s", config.Extensions))
		}
		call = append(call, "--target", "./")

	default:

		return fmt.Errorf("Unknown mta build tool: \"%s\"", config.MtaBuildTool)
	}

	if err = addNpmBinToPath(e); err != nil {
		return err
	}

	log.Entry().Infof("Executing mta build call: \"%s\"", strings.Join(call, " "))

	if err := e.RunExecutable(call[0], call[1:]...); err != nil {
		return err
	}

	commonPipelineEnvironment.mtarFilePath = mtarName
	return nil
}

func getMarJarName(config mtaBuildOptions) string {

	mtaJar := "mta.jar"

	if len(config.MtaJarLocation) > 0 {
		mtaJar = config.MtaJarLocation
	}

	return mtaJar
}

func addNpmBinToPath(e execRunner) error {
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
			return "", err
		}

		if len(mtaID) == 0 {
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
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d\n", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
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

func configureNpmRegistry(registryURI string, registryName string, scope string, e execRunner) error {
	if len(registryURI) == 0 {
		log.Entry().Debugf("No %s npm registry provided via configuration. Leaving npm config untouched.", registryName)
		return nil
	}

	log.Entry().Debugf("Setting %s npm registry to \"%s\"", registryName, registryURI)

	key := "registry"
	if len(scope) > 0 {
		key = fmt.Sprintf("%s:registry", scope)
	}

	if err := e.RunExecutable("npm", "config", "set", key, registryURI); err != nil {
		return err
	}

	return nil
}

func handleSettingsFiles(config mtaBuildOptions,
	p piperutils.FileUtils,
	httpClient piperhttp.Downloader) error {

	if len(config.ProjectSettingsFile) > 0 {

		if err := getSettingsFile(maven.ProjectSettingsFile, config.ProjectSettingsFile, p, httpClient); err != nil {
			return err
		}

	} else {

		log.Entry().Debugf("Project settings file not provided via configuation.")
	}

	if len(config.GlobalSettingsFile) > 0 {

		if err := getSettingsFile(maven.GlobalSettingsFile, config.GlobalSettingsFile, p, httpClient); err != nil {
			return err
		}
	} else {

		log.Entry().Debugf("Global settings file not provided via configuation.")
	}

	return nil
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
