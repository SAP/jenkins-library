package cmd

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gradle"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

var (
	bomGradleTaskName = "cyclonedxBom"
	publishTaskName   = "publish"
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
				artifactId = '{{.ArtifactID}}'
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

const bomInitScriptContent = `
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

type gradleExecuteBuildUtils interface {
	command.ExecRunner
	piperutils.FileUtils
}

type gradleExecuteBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newGradleExecuteBuildUtils() gradleExecuteBuildUtils {
	utils := gradleExecuteBuildUtilsBundle{
		Command: &command.Command{
			StepName: "gradleExecuteBuild",
		},
		Files: &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func gradleExecuteBuild(config gradleExecuteBuildOptions, telemetryData *telemetry.CustomData) {
	utils := newGradleExecuteBuildUtils()
	err := runGradleExecuteBuild(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils) error {
	log.Entry().Info("BOM file creation...")
	if config.CreateBOM {
		if err := createBOM(config, utils); err != nil {
			return err
		}
	}

	// gradle build
	gradleOptions := &gradle.ExecuteOptions{
		BuildGradlePath: config.Path,
		Task:            config.Task,
		UseWrapper:      config.UseWrapper,
	}
	if _, err := gradle.Execute(gradleOptions, utils); err != nil {
		log.Entry().WithError(err).Errorf("gradle build execution was failed: %v", err)
		return err
	}

	log.Entry().Info("Publishing of artifacts to staging repository...")
	if config.Publish {
		if err := publishArtifacts(config, utils); err != nil {
			return err
		}
	}

	return nil
}

func createBOM(config *gradleExecuteBuildOptions, utils gradleExecuteBuildUtils) error {
	gradleOptions := &gradle.ExecuteOptions{
		BuildGradlePath:   config.Path,
		Task:              bomGradleTaskName,
		UseWrapper:        config.UseWrapper,
		InitScriptContent: bomInitScriptContent,
	}
	if _, err := gradle.Execute(gradleOptions, utils); err != nil {
		log.Entry().WithError(err).Errorf("failed to create BOM: %v", err)
		return err
	}
	return nil
}

func publishArtifacts(config *gradleExecuteBuildOptions, utils gradleExecuteBuildUtils) error {
	publishInitScriptContent, err := getPublishInitScriptContent(config)
	if err != nil {
		return fmt.Errorf("failed to get publish init script content: %v", err)
	}
	gradleOptions := &gradle.ExecuteOptions{
		BuildGradlePath:   config.Path,
		Task:              publishTaskName,
		UseWrapper:        config.UseWrapper,
		InitScriptContent: publishInitScriptContent,
	}
	if _, err := gradle.Execute(gradleOptions, utils); err != nil {
		log.Entry().WithError(err).Errorf("failed to publish artifacts: %v", err)
		return err
	}
	return nil
}

func getPublishInitScriptContent(options *gradleExecuteBuildOptions) (string, error) {
	tmpl, err := template.New("resources").Parse(publishInitScriptContentTemplate)
	if err != nil {
		return "", err
	}

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, options)
	if err != nil {
		return "", err
	}

	return generatedCode.String(), nil
}
