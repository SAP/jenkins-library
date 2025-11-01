package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gradle"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	gradleBomFilename        = "bom-gradle"
	stepNameForBuildSettings = "gradleExecuteBuild"
)

var (
	bomGradleTaskName = "cyclonedxBom"
	publishTaskName   = "publish"
	pathToModuleFile  = filepath.Join("build", "publications", "maven", "module.json")
	rootPath          = "."
)

const publishInitScriptContentTemplate = `
{{ if .ApplyPublishingForAllProjects}}allprojects{{else}}rootProject{{ end }} {
    def gradleExecuteBuild_skipPublishingProjects = [{{ if .ApplyPublishingForAllProjects}}{{range .ExcludePublishingForProjects}} "{{.}}",{{end}}{{end}} ];
    if (!gradleExecuteBuild_skipPublishingProjects.contains(project.name)) {
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
                    {{- if .ApplyPublishingForAllProjects }}
                    {{else if .ArtifactID}}
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
}
`

const bomInitScriptContentTemplate = `
initscript {
  repositories {
    mavenCentral()
    maven {
      url "https://plugins.gradle.org/m2/"
    }
  }
  dependencies {
    classpath "org.cyclonedx:cyclonedx-gradle-plugin:1.7.4"
  }
}

allprojects {
    def gradleExecuteBuild_skipBOMProjects = [{{range .ExcludeCreateBOMForProjects}} "{{.}}",{{end}} ];
    if (!gradleExecuteBuild_skipBOMProjects.contains(project.name)) {
        apply plugin: 'java'
        apply plugin: org.cyclonedx.gradle.CycloneDxPlugin

        cyclonedxBom {
            outputName = "` + gradleBomFilename + `"
            outputFormat = "xml"
            schemaVersion = "1.4"
            includeConfigs = ["runtimeClasspath"]
            skipConfigs = ["compileClasspath", "testCompileClasspath"]
        }
    }
}
`

// PublishedArtifacts contains information about published artifacts
type PublishedArtifacts struct {
	Info     Component `json:"component,omitempty"`
	Elements []Element `json:"variants,omitempty"`
}

type Component struct {
	Module string `json:"module,omitempty"`
}

type Element struct {
	Name      string     `json:"name,omitempty"`
	Artifacts []Artifact `json:"files,omitempty"`
}

type Artifact struct {
	Name string `json:"name,omitempty"`
}

type WalkDir func(root string, fn fs.WalkDirFunc) error

type Filepath interface {
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type WalkDirFunc func(root string, fn fs.WalkDirFunc) error

func (f WalkDirFunc) WalkDir(root string, fn fs.WalkDirFunc) error {
	return f(root, fn)
}

type gradleExecuteBuildUtils interface {
	command.ExecRunner
	piperutils.FileUtils
	Filepath
}

type gradleExecuteBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	Filepath
}

func newGradleExecuteBuildUtils() gradleExecuteBuildUtils {
	var walkDirFunc WalkDirFunc = filepath.WalkDir
	utils := gradleExecuteBuildUtilsBundle{
		Command: &command.Command{
			StepName: "gradleExecuteBuild",
		},
		Files:    &piperutils.Files{},
		Filepath: walkDirFunc,
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func gradleExecuteBuild(config gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, pipelineEnv *gradleExecuteBuildCommonPipelineEnvironment) {
	utils := newGradleExecuteBuildUtils()
	err := runGradleExecuteBuild(&config, telemetryData, utils, pipelineEnv)
	if err != nil {
		log.Entry().WithError(err).Fatalf("step execution failed: %v", err)
	}
}

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils, pipelineEnv *gradleExecuteBuildCommonPipelineEnvironment) error {
	log.Entry().Info("BOM file creation...")

	if config.CreateBOM {
		if err := createBOM(config, utils); err != nil {
			return err
		}
	}

	// gradle build
	// if user provides BuildFlags, it is respected over a single Task
	gradleOptions := &gradle.ExecuteOptions{
		BuildGradlePath: config.Path,
		Task:            config.Task,
		BuildFlags:      config.BuildFlags,
		UseWrapper:      config.UseWrapper,
	}
	if _, err := gradle.Execute(gradleOptions, utils); err != nil {
		log.Entry().WithError(err).Errorf("gradle build execution was failed: %v", err)
		return err
	}

	log.Entry().Debugf("creating build settings information...")

	dockerImage, err := GetDockerImageValue(stepNameForBuildSettings)
	if err != nil {
		return fmt.Errorf("failed to retrieve dockerImage configuration: %w", err)
	}

	gradleConfig := buildsettings.BuildOptions{
		CreateBOM:         config.CreateBOM,
		Publish:           config.Publish,
		BuildSettingsInfo: config.BuildSettingsInfo,
		DockerImage:       dockerImage,
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&gradleConfig, stepNameForBuildSettings)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	pipelineEnv.custom.buildSettingsInfo = buildSettingsInfo

	log.Entry().Info("Publishing of artifacts to staging repository...")
	if config.Publish {
		if err := publishArtifacts(config, utils, pipelineEnv); err != nil {
			return err
		}
	}

	return nil
}

func createBOM(config *gradleExecuteBuildOptions, utils gradleExecuteBuildUtils) error {
	createBOMInitScriptContent, err := getInitScriptContent(config, bomInitScriptContentTemplate)
	if err != nil {
		return fmt.Errorf("failed to get BOM init script content: %v", err)
	}
	gradleOptions := &gradle.ExecuteOptions{
		BuildGradlePath:   config.Path,
		Task:              bomGradleTaskName,
		UseWrapper:        config.UseWrapper,
		InitScriptContent: createBOMInitScriptContent,
	}
	if _, err := gradle.Execute(gradleOptions, utils); err != nil {
		log.Entry().WithError(err).Errorf("failed to create BOM: %v", err)
		return err
	}

	// Validate generated SBOMs
	bomFilename := gradleBomFilename + ".xml"
	err = utils.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, bomFilename) {
			log.Entry().Infof("Validating generated SBOM: %s", path)

			if err := piperutils.ValidateCycloneDX14(path); err != nil {
				log.Entry().Warnf("SBOM validation failed: %v", err)
			} else {
				purl := piperutils.GetPurl(path)
				log.Entry().Infof("SBOM validation passed")
				log.Entry().Infof("SBOM PURL: %s", purl)
			}
		}
		return nil
	})
	if err != nil {
		log.Entry().Warnf("Failed to walk directory for SBOM validation: %v", err)
	}

	return nil
}

func publishArtifacts(config *gradleExecuteBuildOptions, utils gradleExecuteBuildUtils, pipelineEnv *gradleExecuteBuildCommonPipelineEnvironment) error {
	publishInitScriptContent, err := getInitScriptContent(config, publishInitScriptContentTemplate)
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
	var artifacts piperenv.Artifacts
	err = utils.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, pathToModuleFile) {
			pathArtifacts, artifactsErr := getPublishedArtifactsNames(path, utils)
			if artifactsErr != nil {
				return fmt.Errorf("failed to get published artifacts in path %s: %v", path, artifactsErr)
			}
			artifacts = append(artifacts, pathArtifacts...)
		}
		return nil
	})
	if err != nil {
		return err
	}
	pipelineEnv.custom.artifacts = artifacts
	return nil
}

func getInitScriptContent(options *gradleExecuteBuildOptions, templateContent string) (string, error) {
	tmpl, err := template.New("resources").Parse(templateContent)
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

func getPublishedArtifactsNames(file string, utils gradleExecuteBuildUtils) (piperenv.Artifacts, error) {
	artifacts := piperenv.Artifacts{}
	publishedArtifacts := PublishedArtifacts{}
	exists, err := utils.FileExists(file)
	if err != nil {
		return nil, fmt.Errorf("failed to check existence of the file '%s': %v", file, err)
	}
	if !exists {
		return nil, fmt.Errorf("failed to get '%s': file does not exist", file)
	}
	content, err := utils.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read '%s': %v", file, err)
	}
	err = json.Unmarshal(content, &publishedArtifacts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal '%s': %v", file, err)
	}
	for _, element := range publishedArtifacts.Elements {
		if element.Name != "apiElements" {
			continue
		}
		for _, artifact := range element.Artifacts {
			artifacts = append(artifacts, piperenv.Artifact{Id: publishedArtifacts.Info.Module, Name: artifact.Name})
		}
	}
	return artifacts, nil
}
