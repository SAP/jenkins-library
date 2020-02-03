package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	defaultNpmRegistry := "npmReg"

	projectSettingsFileSrc := "http://example.org"
	projectSettingsFileDest := "project-settings.xml" // needs to be $HOME/.m2/settings.xml finally
	globalSettingsFileSrc := "http://example.org"
	globalSettingsFileDest := "global-settings.txt" // needs to be $M2_HOME/conf/settings.xml finally
	//

	// project settings file
	if len(projectSettingsFileSrc) > 0 {
		projectSettingsFileParent := filepath.Dir("/home/me/.m2/settings.xml")
		fmt.Printf("ProjectSettingsfileParent: \"%s\"\n", projectSettingsFileParent)
		if strings.HasPrefix(projectSettingsFileSrc, "http:") || strings.HasPrefix(projectSettingsFileSrc, "https:") {
			materialize(projectSettingsFileSrc, projectSettingsFileDest)
		} else {
			piperutils.Copy(projectSettingsFileSrc, projectSettingsFileDest)
		}
	}

	// global settings file
	if len(globalSettingsFileSrc) > 0 {
		materialize(globalSettingsFileSrc, globalSettingsFileDest)
	}

	if len(defaultNpmRegistry) > 0 {
		// REVISIT: would be possible to do this below in the same shell call like the mtar build itself
		s.RunShell("/bin/bash", fmt.Sprintf("npm config set registry %s", defaultNpmRegistry))
	}

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

	// REVISIT: when we have the possibility to provide environment variables from outside we can
	// do the export this way.
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

func materialize(url, file string) error {
	client := &piperhttp.Client{}
	//CHECK:
	// - how does this work with a proxy inbetween?
	// - how does this work with http 302 (relocated) --> curl -L
	response, e := client.SendRequest(http.MethodGet, url, nil, nil, nil)
	if e != nil {
		return e
	}

	if response.StatusCode != 200 {
		fmt.Errorf("Got %d reponse from download attemtp for \"%s\"", response.StatusCode, url)
	}

	body, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return e
	}

	ioutil.WriteFile(file, body, 0644)

	return nil
}
