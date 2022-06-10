package cmd

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"strings"
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
	tasksTaskName                 = "tasks"
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
rootProject {
    ext {
        rootPluginsList = project.hasProperty("rootPluginsList") ? project.getProperty("rootPluginsList") : 'java'
        rootComponent = project.hasProperty("rootComponent") ? project.getProperty("rootComponent") : 'java'
        rootArtifactId = project.hasProperty("rootArtifactId") ? project.getProperty("rootArtifactId") : ''
        rootGroupId = project.hasProperty("rootGroupId") ? project.getProperty("rootGroupId") : ''
        rootVersion = project.hasProperty("rootVersion") ? project.getProperty("rootVersion") : ''
        rootUseDeclaredVersioning = project.hasProperty("rootUseDeclaredVersioning") ? project.getProperty("rootUseDeclaredVersioning").toBoolean() : false
        rootCreateBOM = project.hasProperty("rootCreateBOM") ? project.getProperty("rootCreateBOM").toBoolean() : false
        rootPublish = project.hasProperty("rootPublish") ? project.getProperty("rootPublish").toBoolean() : false
   }

    for(projectPlugin in rootPluginsList.tokenize(",")) {
        apply plugin: projectPlugin
    }

    if (rootCreateBOM) {
        apply plugin: "org.cyclonedx.bom"
    }

    if (rootPublish) {
        publishing {
            publications {
                maven(MavenPublication) {
                    if (!rootUseDeclaredVersioning){
                        versionMapping {
                            usage('java-api') {
                                fromResolutionOf('runtimeClasspath')
                            }
                            usage('java-runtime') {
                                fromResolutionResult()
                            }
                        }
                    }
                    if (rootGroupId != '') {
                        groupId = rootGroupId
                    }
                    if (rootArtifactId != '') {
                        artifactId = rootArtifactId
                    }
                    if (rootVersion != '') {
                        version = rootVersion
                    }
                    from components[rootComponent]
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
subprojects {
    ext {
        subprojectsPluginsList = project.hasProperty("subprojectsPluginsList") ? project.getProperty("subprojectsPluginsList") : 'java'
        subprojectsComponent = project.hasProperty("subprojectsComponent") ? project.getProperty("subprojectsComponent") : 'java'
        subprojectsUseDeclaredVersioning = project.hasProperty("subprojectsUseDeclaredVersioning") ? project.getProperty("subprojectsUseDeclaredVersioning").toBoolean() : false
        subprojectsVersion = project.hasProperty("subprojectsVersion") ? project.getProperty("subprojectsVersion") : ''
        subprojectsGroupId = project.hasProperty("subprojectsGroupId") ? project.getProperty("subprojectsGroupId") : ''
        subprojectsCreateBOM = project.hasProperty("subprojectsCreateBOM") ? project.getProperty("subprojectsCreateBOM").toBoolean() : false
        subprojectsPublish = project.hasProperty("subprojectsPublish") ? project.getProperty("subprojectsPublish").toBoolean() : false

        projectPluginsList = project.hasProperty(project.name+"--pluginsList") ? project.getProperty(project.name+"--pluginsList") : subprojectsPluginsList
        projectPublish = project.hasProperty(project.name+"--publish") ? project.getProperty(project.name+"--publish").toBoolean() : subprojectsPublish
        projectComponent = project.hasProperty(project.name+"--component") ? project.getProperty(project.name+"--component") : subprojectsComponent
        projectUseDeclaredVersioning = project.hasProperty(project.name+"--useDeclaredVersioning") ? project.getProperty(project.name+"--useDeclaredVersioning").toBoolean() : subprojectsUseDeclaredVersioning
        projectVersion = project.hasProperty(project.name+"--version") ? project.getProperty(project.name+"--version") : subprojectsVersion
        projectArtifactId = project.hasProperty(project.name+"--artifactId") ? project.getProperty(project.name+"--artifactId") : ''
        projectGroupId = project.hasProperty(project.name+"--groupId") ? project.getProperty(project.name+"--groupId") : subprojectsGroupId
        projectCreateBOM = project.hasProperty(project.name+"--createBOM") ? project.getProperty(project.name+"--createBOM").toBoolean() : subprojectsCreateBOM
    }

    for(projectPlugin in projectPluginsList.tokenize(",")){
        apply plugin: projectPlugin
    }
    if (projectCreateBOM) {
        apply plugin: "org.cyclonedx.bom"
    }

    if (projectPublish) {
        apply plugin: 'maven-publish'
        publishing{
            publications {
                maven(MavenPublication) {
                    if (!projectUseDeclaredVersioning){
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
`
)

const rootProjectProperties = `
rootPluginsList={{ or .pluginsList "java-library,jacoco"}}
rootComponent={{ or .component "java"}}
rootArtifactId={{ or .artifactId ""}}
rootGroupId={{ or .groupId ""}}
rootVersion={{ or .version ""}}
{{if eq false .createBOM}}rootCreateBOM=false{{end}}{{if .createBOM}}rootCreateBOM={{.createBOM}}{{end}}
{{if eq false .useDeclaredVersioning}}rootUseDeclaredVersioning=false{{end}}{{if .useDeclaredVersioning}}rootUseDeclaredVersioning={{.useDeclaredVersioning}}{{end}}
{{if eq false .publish}}rootPublish=false{{end}}{{if .publish}}rootPublish={{.publish}}{{end}}
`
const subprojectCommonProperties = `
subprojectsPluginsList={{ or .pluginsList "java-library,jacoco"}}
subprojectsComponent={{ or .component "java"}}
subprojectsVersion={{ or .version ""}}
subprojectsGroupId={{ or .groupId ""}}
{{if eq false .publish}}subprojectsPublish=false{{end}}{{if .publish}}subprojectsPublish={{.publish}}{{end}}
{{if eq false .createBOM}}subprojectsCreateBOM=false{{end}}{{if .createBOM}}subprojectsCreateBOM={{.createBOM}}{{end}}
{{if eq false .useDeclaredVersioning}}subprojectsUseDeclaredVersioning=false{{end}}{{if .useDeclaredVersioning}}subprojectsUseDeclaredVersioning={{.useDeclaredVersioning}}{{end}}
`
const subprojectCustomProperties = `
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
	if exists, err := utils.FileExists(oldName); err != nil {
		return errors.Wrapf(err, "unable to check %s file existance", oldName)
	} else {
		if exists {
			if err := utils.FileRename(oldName, newName); err != nil {
				return errors.Wrapf(err, "unable to rename %s file", oldName)
			}
		}
	}
	return nil
}

func safeReadFile(utils gradleExecuteBuildUtils, name string) ([]byte, error) {
	if exists, err := utils.FileExists(name); err != nil {
		return nil, errors.Wrapf(err, "unable to check %s file existance", name)
	} else {
		if exists {
			return utils.FileRead(name)
		}
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
	resultProperties, err := extendProperties(config.GradlePropertiesFile, config.RootProjectConfig, config.SubprojectsCommonConfig, config.SubprojectsCustomConfigs, sensitiveProperties)
	if err != nil {
		return err
	}
	//moving original gradle.properties to tmp file
	if err := safeRenameFile(utils, originalGradlePropertiesFile, temporaryGradlePropertiesFile); err != nil {
		return err
	}
	//then writing generated properties to gradle.properties
	if err := fileUtils.FileWrite(originalGradlePropertiesFile, resultProperties, 0644); err != nil {
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

	if called, err := callGradle(config, utils, initScriptContent, bomGradleTaskName, true); err != nil {
		log.Entry().WithError(err).Errorf("failed to create BOM: %v", err)
		return err
	} else {
		if called {
			log.Entry().Info("BOM file created")
		} else {
			log.Entry().Info("skip BOM file creation")
		}
	}

	if _, err := callGradle(config, utils, "", config.Task, false); err != nil {
		log.Entry().WithError(err).Errorf("gradle %s execution was failed: %v", config.Task, err)
		return err
	}

	if called, err := callGradle(config, utils, initScriptContent, publishTaskName, true); err != nil {
		log.Entry().WithError(err).Errorf("failed to publish: %v", err)
		return err
	} else {
		if called {
			log.Entry().Info("published")
		} else {
			log.Entry().Info("skip publishing")
		}
	}

	return nil
}

func callGradle(config *gradleExecuteBuildOptions, utils gradleExecuteBuildUtils, initScriptContent, task string, optional bool) (bool, error) {
	gradleOptions := &gradle.ExecuteOptions{
		BuildGradlePath:   config.Path,
		UseWrapper:        config.UseWrapper,
		InitScriptContent: initScriptContent,
	}
	if optional {
		gradleOptions.Task = tasksTaskName
		output, err := gradle.Execute(gradleOptions, utils)
		if err != nil {
			return false, err
		}
		if strings.Index(output, task) < 0 {
			return false, nil
		}
	}
	gradleOptions.Task = task
	_, err := gradle.Execute(gradleOptions, utils)
	return true, err
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

func extendProperties(gradlePropertiesFile string, rootProjectConfig map[string]interface{}, subprojectsCommonConfig map[string]interface{}, subprojectsCustomConfigs []map[string]interface{}, sensitiveProperties []byte) ([]byte, error) {
	originalProperties := []byte(``)
	var err error
	if len(gradlePropertiesFile) > 0 {
		exists, err := fileUtils.FileExists(gradlePropertiesFile)
		if err != nil {
			return nil, errors.Wrapf(err, "file '%v' does not exist", gradlePropertiesFile)
		}
		if !exists {
			return nil, errors.Wrapf(err, "file '%v' does not exist", gradlePropertiesFile)
		}
		originalProperties, err = fileUtils.FileRead(gradlePropertiesFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read file '%v'", gradlePropertiesFile)
		}
	}
	sensitiveProperties = append([]byte("\n"), sensitiveProperties...)
	tplRootProps := template.Must(template.New("rootProjectProps").Parse(rootProjectProperties))
	tplSubprojectsCommonProps := template.Must(template.New("subprojectsCommonProps").Parse(subprojectCommonProperties))
	tplSubprojectsCustomProps := template.Must(template.New("subprojectCustomProps").Parse(subprojectCustomProperties))

	properties := append(originalProperties, []byte(sensitiveProperties)...)
	properties, err = appendPropertiesByTemplate(properties, tplRootProps, rootProjectConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate rootProject properties")
	}
	properties, err = appendPropertiesByTemplate(properties, tplSubprojectsCommonProps, subprojectsCommonConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate subprojects common properties")
	}
	for _, cfg := range subprojectsCustomConfigs {
		properties, err = appendPropertiesByTemplate(properties, tplSubprojectsCustomProps, cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate subprojects custom properties")
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
