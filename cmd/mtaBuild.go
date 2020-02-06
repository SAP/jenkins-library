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
	e envExecRunner) error {

	e.Stdout(os.Stderr) // keep stdout clear.
	e.Stderr(os.Stderr)

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
	projectSettingsFileDest := getProjectSettingsFileDest()
	globalSettingsFileSrc := "http://example.org"
	globalSettingsFileDest := getGlobalSettingsFileDest()
	//

	// project settings file
	if len(projectSettingsFileSrc) > 0 {
		if strings.HasPrefix(projectSettingsFileSrc, "http:") || strings.HasPrefix(projectSettingsFileSrc, "https:") {
			materialize(projectSettingsFileSrc, projectSettingsFileDest)
		} else {
			piperutils.Copy(projectSettingsFileSrc, projectSettingsFileDest)
		}
	}

	// global settings file
	if len(globalSettingsFileSrc) > 0 {
		if strings.HasPrefix(projectSettingsFileSrc, "http:") || strings.HasPrefix(projectSettingsFileSrc, "https:") {
			materialize(globalSettingsFileSrc, globalSettingsFileDest)
		} else {
			piperutils.Copy(globalSettingsFileSrc, globalSettingsFileDest)
		}
	}

	if len(defaultNpmRegistry) > 0 {
		// REVISIT: would be possible to do this below in the same shell call like the mtar build itself
		e.RunExecutable("npm", "config", "set", "registry", defaultNpmRegistry)
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
	var call []string

	switch mtaBuildTool {
	case "classic":
		call = append(call, "java", "-jar", mtaJar, fmt.Sprintf("--build-target=%s", buildTarget))
		if len(extensions) != 0 {
			call = append(call, fmt.Sprintf("--extension=%s", extensions))
		}
	case "cloudMbt":
		call = append(call, "mbt", "build", "--platform", platform)
		if len(extensions) != 0 {
			call = append(call, fmt.Sprintf("--extensions=%s", extensions))
		}
		call = append(call, "--target", "./")
	default:
		return fmt.Errorf("Unknown mta build tool: \"${%s}\"", mtaBuildTool)
	}

	log.Entry().Infof("Executing mta build call: \"%s\"", strings.Join(call, " "))

	path := "./node_modules/.bin"
	oldPath := getEnvironmentVariable("PATH")
	if len(oldPath) > 0 {
		path = path + ":" + oldPath
	}
	e.Env(append(os.Environ(), "PATH="+path))

	if err := e.RunExecutable(call[0], strings.Join(call[1:], " ")); err != nil {
		return err
	}

	mtarFilePath := "dummy.mtar"
	commonPipelineEnvironment.mtarFilePath = mtarFilePath
	return nil
}

func getGlobalSettingsFileDest() string {
	return "global-settings.txt" // needs to be $M2_HOME/conf/settings.xml finally
}

func getProjectSettingsFileDest() string {
	return "project-settings.xml" // needs to be $HOME/.m2/settings.xml finally

}

func getEnvironmentVariable(name string) string {

	// in case we have the same name twice we have to take the latest one.
	// hence we reverse the slice in order to get the latest entry first.
	prefix := name+"="
	for _, e := range reverse(os.Environ()) {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

func reverse(s []string) []string {

	// REVISIT: fits better into some string utils

	if len(s) == 0 {
		return s
	}
	return append(reverse(s[1:]), s[0])
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
