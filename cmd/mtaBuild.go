package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"
)

const templateMtaYml = `_schema-version: "2.0.0"
ID: "{{.ID}}"
version: {{.Version}}

parameters:
  hcp-deployer-version: "1.0.0"

modules:
  - name: {{.Name}}
    type: html5
    path: .
    parameters:
       version: {{.Version}}-${timestamp}
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
	e envExecRunner,
	p piperutils.FileUtils,
	httpClient piperhttp.Sender) error {

	e.Stdout(os.Stderr) // keep stdout clear.
	e.Stderr(os.Stderr)

	var err error

	handleSettingsFiles(config, p, httpClient)

	if len(config.DefaultNpmRegistry) > 0 {
		log.Entry().Debugf("Setting default npm registry to \"%s\"", config.DefaultNpmRegistry)
		if err := e.RunExecutable("npm", "config", "set", "registry", config.DefaultNpmRegistry); err != nil {
			return err
		}
	} else {
		log.Entry().Debugf("No default npm registry provided via configuration. Leaving npm config untouched.")
	}

	mtaYamlFile := "mta.yaml"
	mtaYamlFileExists, err := p.FileExists(mtaYamlFile)

	if err != nil {
		return err
	}

	if !mtaYamlFileExists {

		log.Entry().Debugf("mta yaml file not found in project sources.")

		if len(config.ApplicationName) == 0 {
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
			fmt.Errorf("Version not found in \"package.json\" (or wrong type)")
		}

		name, ok := result["name"].(string)
		if !ok {
			fmt.Errorf("Name not found in \"package.json\" (or wrong type)")
		}

		mtaConfig, err := generateMta(name, config.ApplicationName, version)
		if err != nil {
			return err
		}

		p.FileWrite(mtaYamlFile, []byte(mtaConfig), 0644)
		log.Entry().Infof("\"%s\" created.", mtaYamlFile)

	} else {
		log.Entry().Infof("\"%s\" file found in project sources", mtaYamlFile)
	}

	mtaYaml, err := p.FileRead(mtaYamlFile)
	if err != nil {
		return err
	}

	t := time.Now()
	timestamp := fmt.Sprintf("%d%02d%02d%02d%02d%02d\n", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	mtaYamlStr := string(mtaYaml)

	mtaYamlTimestampReplaced := strings.ReplaceAll(mtaYamlStr, "${timestamp}", timestamp)

	if strings.Compare(mtaYamlStr, mtaYamlTimestampReplaced) != 0 {

		if err := p.FileWrite(mtaYamlFile, []byte(mtaYamlTimestampReplaced), 0644); err != nil {
			return err
		}
		log.Entry().Infof("Timestamp replaced in \"%s\"", mtaYamlFile)
	} else {
		log.Entry().Infof("No timestap contained in \"%s\". File has not been modified.", mtaYamlFile)
	}

	var call []string

	mtarName := config.MtarName
	if len(mtarName) == 0 {
		log.Entry().Debugf("mtar name not provided via config. Extracting from file \"%s\"", mtaYamlFile)
		mtaID, err := getMtaID(mtaYamlFile)
		if err != nil {
			return err
		}
		log.Entry().Debugf("mtar name extracted from file \"%s\": \"%s\"", mtaYamlFile, mtaID)
		mtarName = mtaID + ".mtar"
	}

	switch config.MtaBuildTool {
	case "classic":

		mtaJar := "mta.jar"

		if len(config.MtaJarLocation) > 0 {
			mtaJar = config.MtaJarLocation
		}

		buildTarget, err := ValueOfBuildTarget(config.BuildTarget)
		if err != nil {
			return err
		}

		call = append(call, "java", "-jar", mtaJar, "--mtar", mtarName, fmt.Sprintf("--build-target=%s", buildTarget))
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
		return fmt.Errorf("Unknown mta build tool: \"${%s}\"", config.MtaBuildTool)
	}

	log.Entry().Infof("Executing mta build call: \"%s\"", strings.Join(call, " "))

	path := "./node_modules/.bin"
	oldPath := os.Getenv("PATH")
	if len(oldPath) > 0 {
		path = path + ":" + oldPath
	}
	e.Env(append(os.Environ(), "PATH="+path))

	if err := e.RunExecutable(call[0], call[1:]...); err != nil {
		return err
	}

	commonPipelineEnvironment.mtarFilePath = mtarName
	return nil
}

func handleSettingsFiles(config mtaBuildOptions,
	p piperutils.FileUtils,
	httpClient piperhttp.Sender) error {

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

func generateMta(id, name, version string) (string, error) {

	if len(id) == 0 {
		return "", fmt.Errorf("Generating mta file: ID not provided")
	}
	if len(name) == 0 {
		return "", fmt.Errorf("Generating mta file: Name not provided")
	}
	if len(version) == 0 {
		return "", fmt.Errorf("Generating mta file: Version not provided")
	}

	tmpl, e := template.New("mta.yaml").Parse(templateMtaYml)
	if e != nil {
		return "", e
	}

	type properties struct {
		ID      string
		Name    string
		Version string
	}

	props := properties{ID: id, Name: name, Version: version}

	var script bytes.Buffer
	tmpl.Execute(&script, props)
	return script.String(), nil
}

func getMtaID(mtaYamlFile string) (string, error) {

	var result map[string]interface{}
	p, err := ioutil.ReadFile(mtaYamlFile)
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
