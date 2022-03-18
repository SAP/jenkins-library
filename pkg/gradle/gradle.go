package gradle

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	exec                  = "gradle"
	groovyBuildScriptName = "build.gradle"
	kotlinBuildScriptName = "build.gradle.kts"
	publishInitScriptName = "maven-publish.gradle"
)

const publishInitScriptContent = `
rootProject {
    apply plugin: 'maven-publish'
    apply plugin: 'java'

    publishing {
        publications {
            maven(MavenPublication) {
                versionMapping {
                    usage('java-api') {
                        fromResolutionOf('runtimeClasspath')
                    }
                    usage('java-runtime') {
                        fromResolutionResult()
                    }
                }
                groupId = 'org.company'
                artifactId = 'sample'
                version = '1.1'
                from components.java
                pom {
                    name = 'My Library'
                    description = 'A description of my library'
                }
            }
        }
        repositories {
            maven {
                credentials {
                    username = "username"
                    password = "password"
                }
                url = "url"
            }
        }
    }
}
`

type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetEnv(env []string)
	RunExecutable(e string, p ...string) error
	FileExists(filename string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
}

// ExecuteOptions are used by Execute() to construct the Gradle command line.
type ExecuteOptions struct {
	BuildGradlePath    string `json:"path,omitempty"`
	Task               string `json:"task,omitempty"`
	ReturnStdout       bool   `json:"returnStdout,omitempty"`
	Publish            bool   `json:"publish,omitempty"`
	RepositoryURL      string `json:"repositoryUrl,omitempty"`
	RepositoryPassword string `json:"repositoryPassword,omitempty"`
	RepositoryUsername string `json:"repositoryUsername,omitempty"`
}

func Execute(options *ExecuteOptions, utils Utils) (string, error) {
	groovyBuildScriptExists, err := utils.FileExists(filepath.Join(options.BuildGradlePath, groovyBuildScriptName))
	if err != nil {
		return "", fmt.Errorf("failed to check if file exists: %w", err)
	}
	kotlinBuildScriptExists, err := utils.FileExists(filepath.Join(options.BuildGradlePath, kotlinBuildScriptName))
	if err != nil {
		return "", fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !groovyBuildScriptExists && !kotlinBuildScriptExists {
		return "", fmt.Errorf("the specified gradle build script could not be found")
	}

	stdOutBuf, stdOut := evaluateStdOut(options)
	utils.Stdout(stdOut)
	utils.Stderr(log.Writer())

	parameters := getParametersFromOptions(options)

	err = utils.RunExecutable(exec, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{exec}, parameters...)
		return "", fmt.Errorf("failed to run executable, command: '%s', error: %w", commandLine, err)
	}

	if options.Publish {
		if err := publish(options, utils); err != nil {
			return "", fmt.Errorf("failed to create BOM: %w", err)
		}
	}

	if stdOutBuf == nil {
		return "", nil
	}
	return string(stdOutBuf.Bytes()), nil
}

func evaluateStdOut(options *ExecuteOptions) (*bytes.Buffer, io.Writer) {
	var stdOutBuf *bytes.Buffer
	stdOut := log.Writer()
	if options.ReturnStdout {
		stdOutBuf = new(bytes.Buffer)
		stdOut = io.MultiWriter(stdOut, stdOutBuf)
	}
	return stdOutBuf, stdOut
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

func publish(options *ExecuteOptions, utils Utils) error {
	err := utils.FileWrite(filepath.Join(options.BuildGradlePath, publishInitScriptName), []byte(publishInitScriptContent), 0644)
	if err != nil {
		return fmt.Errorf("failed create init script: %w", err)
	}
	defer utils.FileRemove(filepath.Join(options.BuildGradlePath, publishInitScriptName))
	if err := utils.RunExecutable(exec, "--init-script", filepath.Join(options.BuildGradlePath, publishInitScriptName), "publish"); err != nil {
		return fmt.Errorf("publishing failed: %w", err)
	}

	return nil
}
