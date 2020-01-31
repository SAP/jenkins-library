package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
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

func mtaBuild(config mtaBuildOptions, commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment) error {
	log.Entry().Info("Launching mta build")
	return runMtaBuild(config, commonPipelineEnvironment, &command.Command{})
}

func runMtaBuild(config mtaBuildOptions, commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment,
	s shellRunner) error {

	s.Stdout(os.Stderr) // keep stdout clear.
	s.Stderr(os.Stderr)

	//
	//mtaBuildTool := "classic"
	mtaBuildTool := "cloudMbt"
	buildTarget := "buildTarget"
	extensions := "ext"
	platform := "platform"
	//	applicationName := ""
	applicationName := "myApp"
	//

	mtaYamlFile := "mta.yaml"
	mtaYamlFileExists, err := piperutils.FileExists(mtaYamlFile)

	if err != nil {
		return err
	}

	if !mtaYamlFileExists {

		if len(applicationName) == 0 {
			return fmt.Errorf("'%[1]s' not found in project sources and 'applicationName' not provided as parameter - cannot generate '%[1]s' file", mtaYamlFile)
		}

		mtaConfig, err := generateMta("myID", applicationName, "myVersion")
		if err != nil {
			return err
		}

		// todo prepare for mocking
		ioutil.WriteFile(mtaYamlFile, []byte(mtaConfig), 0644)
		log.Entry().Infof("\"%s\" created.", mtaYamlFile)

	} else {
		log.Entry().Infof("\"%s\" file found in project sources", mtaYamlFile)
	}

	var mtaJar = "mta.jar"
	var mtaCall = `Echo "Hello MTA"`
	var options = []string{}


	switch mtaBuildTool {
	case "classic":
		options = append(options, fmt.Sprintf("--build-target=%s", buildTarget))
		if len(extensions) != 0 {
			options = append(options, fmt.Sprintf("--extension=%s", extensions))
		}
		mtaCall = fmt.Sprintf("java -jar %s %s build", mtaJar, strings.Join(options, " "))
	case "cloudMbt":
		options = append(options, fmt.Sprintf("--platform %s", platform))
		if len(extensions) != 0 {
			options = append(options, fmt.Sprintf("--extensions=%s", extensions))
		}
		options = append(options, "--target ./")
		mtaCall = fmt.Sprintf("mbt build %s", strings.Join(options, " "))
	default:
		return fmt.Errorf("Unknown mta build tool: \"${%s}\"", mtaBuildTool)
	}

	log.Entry().Infof("Executing mta build call: \"%s\"", mtaCall)

	script := fmt.Sprintf(`#!/bin/bash
	export PATH=./node_modules/.bin:$PATH
	echo "[DEBUG] PATH: ${PATH}"
	%s`, mtaCall)

	if err := s.RunShell("/bin/bash", script); err != nil {
		return err
	}

	mtarFilePath := "dummy.mtar"
	commonPipelineEnvironment.mtarFilePath = mtarFilePath
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
