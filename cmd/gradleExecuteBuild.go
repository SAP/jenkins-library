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
	bomGradleTaskName         = "cyclonedxBom"
	publishTaskName           = "publish"
	tasksTaskName             = "tasks"
	initScriptContentTemplate = `
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

func runGradleExecuteBuild(config *gradleExecuteBuildOptions, telemetryData *telemetry.CustomData, utils gradleExecuteBuildUtils) error {
	err := extendProperties(config.GradlePropertiesFile, config.RootProjectConfig, config.SubprojectsCommonConfig, config.SubprojectsCustomConfigs, config.GradleSensitivePropertiesFile)
	if err != nil {
		return err
	}
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

func extendProperties(gradlePropertiesFile string, rootProjectConfig map[string]interface{}, subprojectsCommonConfig map[string]interface{}, subprojectsCustomConfigs []map[string]interface{}, sensitiveProperties string) error {
	properties := []byte(``)
	var err error
	if len(gradlePropertiesFile) > 0 {
		exists, err := fileUtils.FileExists(gradlePropertiesFile)
		if err != nil {
			return errors.Wrapf(err, "file '%v' does not exist", gradlePropertiesFile)
		}
		if exists {
			properties, err = fileUtils.FileRead(gradlePropertiesFile)
			if err != nil {
				return errors.Wrapf(err, "failed to read file '%v'", gradlePropertiesFile)
			}
		}
	}
	sensitiveProperties = "\n" + sensitiveProperties
	tplRootProps := template.Must(template.New("rootProjectProps").Parse(rootProjectProperties))
	tplSubprojectsCommonProps := template.Must(template.New("subprojectsCommonProps").Parse(subprojectCommonProperties))
	tplSubprojectsCustomProps := template.Must(template.New("subprojectCustomProps").Parse(subprojectCustomProperties))

	properties = append(properties, []byte(sensitiveProperties)...)
	properties, err = appendPropertiesByTemplate(properties, tplRootProps, rootProjectConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to generate rootProject properties")
	}
	properties, err = appendPropertiesByTemplate(properties, tplSubprojectsCommonProps, subprojectsCommonConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to generate subprojects common properties")
	}
	for _, cfg := range subprojectsCustomConfigs {
		properties, err = appendPropertiesByTemplate(properties, tplSubprojectsCustomProps, cfg)
		if err != nil {
			return errors.Wrapf(err, "failed to generate subprojects custom properties")
		}
	}

	resultGradlePropertiesFile := "gradle.properties"
	if err := fileUtils.FileWrite(resultGradlePropertiesFile, properties, 0644); err != nil {
		return errors.Wrapf(err, "failed to read file '%v'", gradlePropertiesFile)
	}
	return nil
}

func appendPropertiesByTemplate(source []byte, tpl *template.Template, config map[string]interface{}) ([]byte, error) {
	generatedProps := bytes.Buffer{}
	err := tpl.Execute(&generatedProps, config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate %s properties", tpl.Name())
	}
	return append(source, generatedProps.Bytes()...), nil
}
