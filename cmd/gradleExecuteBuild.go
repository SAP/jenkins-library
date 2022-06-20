package cmd

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gradle"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	originalGradlePropertiesFile  = "gradle.properties"
	temporaryGradlePropertiesFile = "gradle.properties.tmp"
	bomGradleTaskName             = "cyclonedxBom"
	publishTaskName               = "publish"
	initScriptContentTemplate     = `
initscript {
  repositories {
    mavenCentral()
    maven {
      url "https://plugins.gradle.org/m2/"
    }
    maven {
      url 'https://oss.sonatype.org/content/repositories/snapshots'
    }
  }
  dependencies {
    classpath "com.cyclonedx:cyclonedx-gradle-plugin:1.5.0"
  }
}

apply plugin: EnterpriseRepositoryPlugin
class EnterpriseRepositoryPlugin implements Plugin < Gradle > {

  void apply(Gradle gradle) {

    gradle.allprojects {
      project ->
      ext {
        projectsPluginsList = properties.hasProperty("projectsPluginsList") ? properties.getProperty("projectsPluginsList") : 'java'
        projectsComponent = properties.hasProperty("projectsComponent") ? properties.getProperty("projectsComponent") : 'java'
        projectsUseDeclaredVersioning = project.hasProperty("projectsUseDeclaredVersioning") ? project.getProperty("projectsUseDeclaredVersioning").toBoolean() : false
        projectsVersion = project.hasProperty("projectsVersion") ? project.getProperty("projectsVersion") : ''
        projectsGroupId = project.hasProperty("projectsGroupId") ? project.getProperty("projectsGroupId") : ''
        projectsCreateBOM = project.hasProperty("projectsCreateBOM") ? project.getProperty("projectsCreateBOM").toBoolean() : false
        projectsPublish = project.hasProperty("projectsPublish") ? project.getProperty("projectsPublish").toBoolean() : false

        projectPluginsList = project.hasProperty(project.name + "--pluginsList") ? project.getProperty(project.name + "--pluginsList") : projectsPluginsList
        projectPublish = project.hasProperty(project.name + "--publish") ? project.getProperty(project.name + "--publish").toBoolean() : projectsPublish
        projectComponent = project.hasProperty(project.name + "--component") ? project.getProperty(project.name + "--component") : projectsComponent
        projectUseDeclaredVersioning = project.hasProperty(project.name + "--useDeclaredVersioning") ? project.getProperty(project.name + "--useDeclaredVersioning").toBoolean() : projectsUseDeclaredVersioning
        projectVersion = project.hasProperty(project.name + "--version") ? project.getProperty(project.name + "--version") : projectsVersion
        projectArtifactId = project.hasProperty(project.name + "--artifactId") ? project.getProperty(project.name + "--artifactId") : ''
        projectGroupId = project.hasProperty(project.name + "--groupId") ? project.getProperty(project.name + "--groupId") : projectsGroupId
        projectCreateBOM = project.hasProperty(project.name + "--createBOM") ? project.getProperty(project.name + "--createBOM").toBoolean() : projectsCreateBOM
      }

      for (projectPlugin in projectPluginsList.tokenize(",")) {
        apply plugin: projectPlugin
      }
      if (projectCreateBOM) {
        apply plugin: org.cyclonedx.gradle.CycloneDxPlugin
      }

      if (projectPublish) {
        apply plugin: 'maven-publish'
        publishing {
          publications {
            maven(MavenPublication) {
              if (!projectUseDeclaredVersioning) {
                versionMapping {
                  usage('java-api') {
                    fromResolutionOf('runtimeClasspath')
                  }
                  usage('java-runtime') {
                    fromResolutionResult()
                  }
                }
              }
              if (projectArtifactId != '') {
                groupId = projectGroupId
              }
              if (projectArtifactId != '') {
                artifactId = projectArtifactId
              }
              if (projectVersion != '') {
                version = projectVersion
              }
              from components[projectComponent]
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
  }
}
`
)

const projectCommonProperties = `
projectsPluginsList={{ or .pluginsList "java-library,jacoco"}}
projectsComponent={{ or .component "java"}}
projectsVersion={{ or .version ""}}
projectsGroupId={{ or .groupId ""}}
{{if eq false .publish}}projectsPublish=false{{end}}{{if .publish}}projectsPublish={{.publish}}{{end}}
{{if eq false .createBOM}}projectsCreateBOM=false{{end}}{{if .createBOM}}projectsCreateBOM={{.createBOM}}{{end}}
{{if eq false .useDeclaredVersioning}}projectsUseDeclaredVersioning=false{{end}}{{if .useDeclaredVersioning}}projectsUseDeclaredVersioning={{.useDeclaredVersioning}}{{end}}
`
const projectCustomProperties = `
{{.projectName}}--pluginsList={{or .pluginsList ""}}
{{.projectName}}--component={{or .component ""}}
{{.projectName}}--version={{or .version ""}}
{{.projectName}}--artifactId={{or .artifactId ""}}
{{.projectName}}--groupId={{or .groupId ""}}
{{if eq false .publish}}{{.projectName}}--publish=false{{end}}{{if .publish}}{{.projectName}}--publish={{.publish}}{{end}}
{{if eq false .createBOM}}{{.projectName}}--createBOM=false{{end}}{{if .createBOM}}{{.projectName}}--createBOM={{.createBOM}}{{end}}
{{if eq false .useDeclaredVersioning}}{{.projectName}}--useDeclaredVersioning=false{{end}}{{if .useDeclaredVersioning}}{{.projectName}}--useDeclaredVersioning={{.useDeclaredVersioning}}{{end}}
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
		Command: &command.Command{},
		Files:   &piperutils.Files{},
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

func safeRenameFile(utils gradleExecuteBuildUtils, oldName, newName string) error {
	exists, err := utils.FileExists(oldName)
	if err != nil {
		return errors.Wrapf(err, "unable to check %s file existance", oldName)
	}
	if exists {
		if err := utils.FileRename(oldName, newName); err != nil {
			return errors.Wrapf(err, "unable to rename %s file", oldName)
		}
	}
	return nil
}

func safeReadFile(utils gradleExecuteBuildUtils, name string) ([]byte, error) {
	exists, err := utils.FileExists(name)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to check %s file existance", name)
	}
	if exists {
		return utils.FileRead(name)
	}
	return nil, nil
}

func safeRemoveFile(utils gradleExecuteBuildUtils, name string) error {
	if exists, err := utils.FileExists(name); err != nil {
		log.Entry().WithError(err).Errorf("unable to check %s file existance", name)
	} else {
		if exists {
			if err := utils.FileRemove(name); err != nil {
				log.Entry().WithError(err).Errorf("unable to remove %s file", name)
			}
		}
	}
	return nil
}

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils) error {
	sensitiveProperties, err := safeReadFile(utils, config.GradleSensitivePropertiesFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read file '%v'", config.GradleSensitivePropertiesFile)
	}
	originalProperties, err := getOriginalProperties(utils, config.GradlePropertiesFile)
	if err != nil {
		return err
	}
	resultProperties, err := extendProperties(originalProperties, config.ProjectsCommonConfig, config.ProjectsCustomConfigs, sensitiveProperties)
	if err != nil {
		return err
	}
	//moving original gradle.properties to tmp file
	if err := safeRenameFile(utils, originalGradlePropertiesFile, temporaryGradlePropertiesFile); err != nil {
		return err
	}
	//then writing generated properties to gradle.properties
	if err := utils.FileWrite(originalGradlePropertiesFile, resultProperties, 0644); err != nil {
		return errors.Wrapf(err, "failed to read file '%v'", originalGradlePropertiesFile)
	}
	//once done - removing generated properties and renaming origina gradle.properties back from tmp file
	defer func() {
		if err := safeRemoveFile(utils, originalGradlePropertiesFile); err != nil {
			log.Entry().Error(err)
		}
		if err := safeRenameFile(utils, temporaryGradlePropertiesFile, originalGradlePropertiesFile); err != nil {
			log.Entry().Error(err)
		}
	}()

	initScriptContent, err := getInitScript(config)
	if err != nil {
		return fmt.Errorf("failed to get publish init script content: %v", err)
	}

	gradleOptions := &gradle.ExecuteOptions{
		BuildGradlePath:   config.Path,
		UseWrapper:        config.UseWrapper,
		InitScriptContent: initScriptContent,
		InitScriptTasks:   []string{bomGradleTaskName, publishTaskName},
		Tasks:             config.Tasks,
		SkipTasks:         config.SkipTasks,
	}
	_, err = gradle.Execute(gradleOptions, utils)
	return err
}

func getInitScript(options *gradleExecuteBuildOptions) (string, error) {
	tmpl, err := template.New("resources").Parse(initScriptContentTemplate)
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

func getOriginalProperties(utils gradleExecuteBuildUtils, gradlePropertiesFile string) ([]byte, error) {
	originalProperties := []byte(``)
	if len(gradlePropertiesFile) == 0 {
		return originalProperties, nil
	}
	exists, err := utils.FileExists(gradlePropertiesFile)
	if err != nil {
		return nil, errors.Wrapf(err, "file '%v' does not exist", gradlePropertiesFile)
	}
	if !exists {
		return nil, errors.Wrapf(err, "file '%v' does not exist", gradlePropertiesFile)
	}
	originalProperties, err = utils.FileRead(gradlePropertiesFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file '%v'", gradlePropertiesFile)
	}
	return originalProperties, nil
}

func extendProperties(originalProperties []byte, projectsCommonConfig map[string]interface{}, projectsCustomConfigs []map[string]interface{}, sensitiveProperties []byte) ([]byte, error) {
	sensitiveProperties = append([]byte("\n"), sensitiveProperties...)
	tplProjectsCommonProps := template.Must(template.New("projectsCommonProps").Parse(projectCommonProperties))
	tplProjectsCustomProps := template.Must(template.New("projectCustomProps").Parse(projectCustomProperties))

	properties := append(originalProperties, sensitiveProperties...)
	properties, err := appendPropertiesByTemplate(properties, tplProjectsCommonProps, projectsCommonConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate projects common properties")
	}
	for _, cfg := range projectsCustomConfigs {
		properties, err = appendPropertiesByTemplate(properties, tplProjectsCustomProps, cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate projects custom properties")
		}
	}
	return properties, nil
}

func appendPropertiesByTemplate(source []byte, tpl *template.Template, config map[string]interface{}) ([]byte, error) {
	generatedProps := bytes.Buffer{}
	err := tpl.Execute(&generatedProps, config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate %s properties", tpl.Name())
	}
	return append(source, generatedProps.Bytes()...), nil
}
