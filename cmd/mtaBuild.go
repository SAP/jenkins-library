package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
		return -1, fmt.Errorf("Unknown BuildTarget: '%s'", str)
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

func mtaBuild(config mtaBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment) error {
	log.Entry().Info("Launching mta build")
	return runMtaBuild(config, commonPipelineEnvironment, &command.Command{})
}

func runMtaBuild(config mtaBuildOptions, commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment,
	e envExecRunner) error {

	e.Stdout(os.Stderr) // keep stdout clear.
	e.Stderr(os.Stderr)

	projectSettingsFileSrc := "http://example.org"
	projectSettingsFileDest, err := getProjectSettingsFileDest()
	if err != nil {
		return err
	}
	globalSettingsFileSrc := "http://example.org"
	globalSettingsFileDest, err := getGlobalSettingsFileDest()
	if err != nil {
		return err
	}
	//

	if err = materialize(projectSettingsFileSrc, projectSettingsFileDest); err != nil {
		return err
	}

	if err = materialize(globalSettingsFileSrc, globalSettingsFileDest); err != nil {
		return err
	}

	if len(config.DefaultNpmRegistry) > 0 {
		if err := e.RunExecutable("npm", "config", "set", "registry", config.DefaultNpmRegistry); err != nil {
			return err
		}
	}

	mtaYamlFile := "mta.yaml"
	mtaYamlFileExists, err := piperutils.FileExists(mtaYamlFile)

	if err != nil {
		return err
	}

	if !mtaYamlFileExists {

		if len(config.ApplicationName) == 0 {
			return fmt.Errorf("'%[1]s' not found in project sources and 'applicationName' not provided as parameter - cannot generate '%[1]s' file", mtaYamlFile)
		}

		mtaConfig, err := generateMta("myID", config.ApplicationName, "myVersion")
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

	switch config.MtaBuildTool {
	case "classic":

		buildTarget, err := ValueOfBuildTarget(config.BuildTarget)
		if err != nil {
			return err
		}

		call = append(call, "java", "-jar", mtaJar, fmt.Sprintf("--build-target=%s", buildTarget))
		if len(config.Extensions) != 0 {
			call = append(call, fmt.Sprintf("--extension=%s", config.Extensions))
		}
	case "cloudMbt":

		platform, err := ValueOfBuildTarget(config.Platform)
		if err != nil {
			return err
		}

		call = append(call, "mbt", "build", "--platform", platform.String())
		if len(config.Extensions) != 0 {
			call = append(call, fmt.Sprintf("--extensions=%s", config.Extensions))
		}
		call = append(call, "--target", "./")
	default:
		return fmt.Errorf("Unknown mta build tool: \"${%s}\"", config.MtaBuildTool)
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

func getGlobalSettingsFileDest() (string, error) {

	m2Home := getEnvironmentVariable("M2_HOME")

	if len(m2Home) == 0 {
		return "", errors.New("Environment variable \"M2_HOME\" not set or empty")
	}
	return m2Home + "/conf/settings.xml", nil
}

func getProjectSettingsFileDest() (string, error) {

	home := getEnvironmentVariable("HOME")

	if len(home) == 0 {
		return "", errors.New("Environment variable \"HOME\" not set or empty")
	}
	return home + "/.m2/settings.xml", nil
}

func getEnvironmentVariable(name string) string {

	// in case we have the same name twice we have to take the latest one.
	// hence we reverse the slice in order to get the latest entry first.
	prefix := name + "="
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

func materialize(src, dest string) error {

	if len(src) > 0 {

		log.Entry().Debugf("Copying file \"%s\" to \"%s\"", src, dest)

		if strings.HasPrefix(src, "http:") || strings.HasPrefix(src, "https:") {
			if err := materializeURL(src, dest); err != nil {
				return err
			}
		} else {

			parent := filepath.Dir(dest)

			exists, err := piperutils.FileExists(parent)

			if err != nil {
				return err
			}

			if !exists {
				if err = os.MkdirAll(parent, 0664); err != nil {
					return err
				}
			}

			if _, err := piperutils.Copy(src, dest); err != nil {
				return err
			}
		}
	}
	log.Entry().Debugf("File \"%s\" copied to \"%s\"", src, dest)
	return nil
}

func materializeURL(url, file string) error {

	var e error
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

	e = ioutil.WriteFile(file, body, 0644)
	if e != nil {
		return e
	}

	return nil
}
