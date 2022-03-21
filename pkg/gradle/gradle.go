package gradle

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	exec                  = "gradle"
	bomTaskName           = "cyclonedxBom"
	groovyBuildScriptName = "build.gradle"
	kotlinBuildScriptName = "build.gradle.kts"
	initScriptName        = "cyclonedx.gradle"
)

const initScriptContent = `
initscript {
  repositories {
    mavenCentral()
    maven {
      url "https://plugins.gradle.org/m2/"
    }
  }
  dependencies {
    classpath "com.cyclonedx:cyclonedx-gradle-plugin:1.5.0"
  }
}

rootProject {
    apply plugin: 'java'
    apply plugin: 'maven'
    apply plugin: org.cyclonedx.gradle.CycloneDxPlugin
}
`

type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	FileExists(filename string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
}

// ExecuteOptions are used by Execute() to construct the Gradle command line.
type ExecuteOptions struct {
	BuildGradlePath string `json:"path,omitempty"`
	Task            string `json:"task,omitempty"`
	CreateBOM       bool   `json:"createBOM,omitempty"`
}

func Execute(options *ExecuteOptions, utils Utils) error {
	groovyBuildScriptExists, err := utils.FileExists(filepath.Join(options.BuildGradlePath, groovyBuildScriptName))
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	kotlinBuildScriptExists, err := utils.FileExists(filepath.Join(options.BuildGradlePath, kotlinBuildScriptName))
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !groovyBuildScriptExists && !kotlinBuildScriptExists {
		return fmt.Errorf("the specified gradle build script could not be found")
	}

	if options.CreateBOM {
		if err := createBOM(options, utils); err != nil {
			return fmt.Errorf("failed to create BOM: %w", err)
		}
	}

	parameters := getParametersFromOptions(options)

	err = utils.RunExecutable(exec, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{exec}, parameters...)
		return fmt.Errorf("failed to run executable, command: '%s', error: %w", commandLine, err)
	}

	return nil
}

func getParametersFromOptions(options *ExecuteOptions) []string {
	var parameters []string

	// default value for task is 'build', so no necessary to checking for empty parameter
	parameters = append(parameters, options.Task)

	// resolve path for build.gradle execution
	if options.BuildGradlePath != "" {
		parameters = append(parameters, "-p")
		parameters = append(parameters, options.BuildGradlePath)
	}

	return parameters
}

// CreateBOM generates BOM file using CycloneDX
func createBOM(options *ExecuteOptions, utils Utils) error {
	// check if gradle task cyclonedxBom exists
	stdOutBuf := new(bytes.Buffer)
	stdOut := log.Writer()
	stdOut = io.MultiWriter(stdOut, stdOutBuf)
	utils.Stdout(stdOut)
	if err := utils.RunExecutable(exec, "tasks"); err != nil {
		return fmt.Errorf("failed list gradle tasks: %w", err)
	}
	if strings.Contains(stdOutBuf.String(), bomTaskName) {
		if err := utils.RunExecutable(exec, bomTaskName); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	} else {
		err := utils.FileWrite(filepath.Join(options.BuildGradlePath, initScriptName), []byte(initScriptContent), 0644)
		if err != nil {
			return fmt.Errorf("failed create init script: %w", err)
		}
		defer utils.FileRemove(filepath.Join(options.BuildGradlePath, initScriptName))
		if err := utils.RunExecutable(exec, "--init-script", filepath.Join(options.BuildGradlePath, initScriptName), bomTaskName); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	return nil
}
