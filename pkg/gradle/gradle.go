package gradle

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

const (
	exec                  = "gradle"
	bomTaskName           = "cyclonedxBom"
	groovyBuildScriptName = "build.gradle"
	kotlinBuildScriptName = "build.gradle.kts"
	createBOMScriptName   = "cyclonedx.gradle"
	publishInitScriptName = "maven-publish.gradle"
)

const publishInitScriptContentTemplate = `
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
				{{- if .ArtifactGroupID}}
				groupId = '{{.ArtifactGroupID}}'
				{{- end }}
				{{- if .ArtifactID}}
				groupId = '{{.ArtifactID}}'
				{{- end }}
				{{- if .ArtifactVersion}}
				version = '{{.ArtifactVersion}}'
				{{- end }}
                from components.java
            }
        }
        repositories {
            maven {
                credentials {
                    username = "{{.RepositoryUsername}}"
                    password = "{{.RepositoryPassword}}"
                }
                url = "{{.RepositoryURL}}"
            }
        }
    }
}
`

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
	command.ExecRunner
	piperutils.FileUtils
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

// ExecuteOptions are used by Execute() to construct the Gradle command line.
type ExecuteOptions struct {
	BuildGradlePath    string `json:"path,omitempty"`
	Task               string `json:"task,omitempty"`
	CreateBOM          bool   `json:"createBOM,omitempty"`
	ReturnStdout       bool   `json:"returnStdout,omitempty"`
	Publish            bool   `json:"publish,omitempty"`
	ArtifactVersion    string `json:"artifactVersion,omitempty"`
	ArtifactGroupID    string `json:"artifactGroupId,omitempty"`
	ArtifactID         string `json:"artifactId,omitempty"`
	RepositoryURL      string `json:"repositoryUrl,omitempty"`
	RepositoryPassword string `json:"repositoryPassword,omitempty"`
	RepositoryUsername string `json:"repositoryUsername,omitempty"`
}

func Execute(options *ExecuteOptions, utils Utils) error {
	groovyBuildScriptExists, err := utils.FileExists(filepath.Join(options.BuildGradlePath, groovyBuildScriptName))
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %v", err)
	}
	kotlinBuildScriptExists, err := utils.FileExists(filepath.Join(options.BuildGradlePath, kotlinBuildScriptName))
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %v", err)
	}
	if !groovyBuildScriptExists && !kotlinBuildScriptExists {
		return fmt.Errorf("the specified gradle build script could not be found")
	}

	if options.CreateBOM {
		if err := createBOM(options, utils); err != nil {
			return fmt.Errorf("failed to create BOM: %v", err)
		}
	}

	parameters := getParametersFromOptions(options)

	err = utils.RunExecutable(exec, parameters...)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		commandLine := append([]string{exec}, parameters...)
		return fmt.Errorf("failed to run executable, command: '%s', error: %v", commandLine, err)
	}

	if options.Publish {
		if err := publish(options, utils); err != nil {
			return fmt.Errorf("failed to publish artifacts to staging repository: %v", err)
		}
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

func publish(options *ExecuteOptions, utils Utils) error {
	log.Entry().Info("Publishing artifact to staging repository...")
	if len(options.RepositoryURL) == 0 {
		return fmt.Errorf("there's no target repository for binary publishing configured")
	}
	if len(options.ArtifactVersion) == 0 {
		artifactOpts := versioning.Options{
			VersioningScheme: "library",
		}

		artifact, err := versioning.GetArtifact("gradle", "", &artifactOpts, utils)

		if err != nil {
			return err
		}

		options.ArtifactVersion, err = artifact.GetVersion()

		if err != nil {
			return err
		}
	}
	publishInitScriptContent, err := getPublishInitScriptContent(options)
	if err != nil {
		return fmt.Errorf("failed to get init script content: %v", err)
	}
	err = utils.FileWrite(filepath.Join(options.BuildGradlePath, publishInitScriptName), []byte(publishInitScriptContent), 0644)
	if err != nil {
		return fmt.Errorf("failed create init script: %v", err)
	}
	// defer utils.FileRemove(filepath.Join(options.BuildGradlePath, publishInitScriptName))

	if err := utils.RunExecutable(exec, "--init-script", filepath.Join(options.BuildGradlePath, publishInitScriptName), "--info", "publish"); err != nil {
		return fmt.Errorf("publishing failed: %v", err)
	}
	return nil
}

func getPublishInitScriptContent(options *ExecuteOptions) (string, error) {
	tmpl, err := template.New("resources").Parse(publishInitScriptContentTemplate)
	if err != nil {
		return "", err
	}

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, options)
	if err != nil {
		return "", err
	}

	return string(generatedCode.Bytes()), nil
}

// CreateBOM generates BOM file using CycloneDX
func createBOM(options *ExecuteOptions, utils Utils) error {
	log.Entry().Info("BOM creation...")
	// check if gradle task cyclonedxBom exists
	stdOutBuf := new(bytes.Buffer)
	stdOut := log.Writer()
	stdOut = io.MultiWriter(stdOut, stdOutBuf)
	utils.Stdout(stdOut)
	if err := utils.RunExecutable(exec, "tasks"); err != nil {
		return fmt.Errorf("failed list gradle tasks: %v", err)
	}
	if strings.Contains(stdOutBuf.String(), bomTaskName) {
		if err := utils.RunExecutable(exec, bomTaskName); err != nil {
			return fmt.Errorf("BOM creation failed: %v", err)
		}
	} else {
		err := utils.FileWrite(filepath.Join(options.BuildGradlePath, createBOMScriptName), []byte(initScriptContent), 0644)
		if err != nil {
			return fmt.Errorf("failed create init script: %v", err)
		}
		defer utils.FileRemove(filepath.Join(options.BuildGradlePath, createBOMScriptName))
		if err := utils.RunExecutable(exec, "--init-script", filepath.Join(options.BuildGradlePath, createBOMScriptName), bomTaskName); err != nil {
			return fmt.Errorf("BOM creation failed: %v", err)
		}
	}

	return nil
}
