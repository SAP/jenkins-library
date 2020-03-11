package cmd

import (
	"errors"
	"io/ioutil"
	"strings"

	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
)

// nexusUploadUtils defines an interface for utility functionality used from external packages,
// so it can be easily mocked for testing.
type nexusUploadUtils interface {
	fileExists(path string) (bool, error)
	fileRead(path string) ([]byte, error)
	usesMta() bool
	usesMaven() bool
	getEnvParameter(path, name string) string
	evaluateProperty(pomFile, expression string) (string, error)
}

type utilsBundle struct {
	projectStructure piperutils.ProjectStructure
	fileUtils        piperutils.Files
}

func newUtilsBundle() *utilsBundle {
	return &utilsBundle{
		projectStructure: piperutils.ProjectStructure{},
		fileUtils:        piperutils.Files{},
	}
}

func (u *utilsBundle) fileExists(path string) (bool, error) {
	return u.fileUtils.FileExists(path)
}

func (u *utilsBundle) fileRead(path string) ([]byte, error) {
	return u.fileUtils.FileRead(path)
}

func (u *utilsBundle) usesMta() bool {
	return u.projectStructure.UsesMta()
}

func (u *utilsBundle) usesMaven() bool {
	return u.projectStructure.UsesMaven()
}

func (u *utilsBundle) getEnvParameter(path, name string) string {
	return piperenv.GetParameter(path, name)
}

func (u *utilsBundle) evaluateProperty(pomFile, expression string) (string, error) {
	execRunner := command.Command{}
	execRunner.Stdout(ioutil.Discard)
	execRunner.Stderr(ioutil.Discard)

	expressionDefine := "-Dexpression=" + expression

	options := maven.ExecuteOptions{
		PomPath:      pomFile,
		Goals:        []string{"org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"},
		Defines:      []string{expressionDefine, "-DforceStdout", "-q"},
		ReturnStdout: true,
	}
	value, err := maven.Execute(&options, &execRunner)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(value, "null object or invalid expression") {
		return "", fmt.Errorf("expression '%s' in file '%s' could not be resolved", expression, pomFile)
	}
	log.Entry().Debugf("Evaluated expression '%s' in file '%s' as '%s'\n", expression, pomFile, value)
	return value, nil
}

func nexusUpload(options nexusUploadOptions, _ *telemetry.CustomData) {
	utils := newUtilsBundle()
	uploader := nexus.Upload{}

	err := runNexusUpload(utils, &uploader, &options)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNexusUpload(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	if utils.usesMta() {
		log.Entry().Info("MTA project structure detected")
		return uploadMTA(utils, uploader, options)
	} else if utils.usesMaven() {
		log.Entry().Info("Maven project structure detected")
		return uploadMaven(uploader, options)
	} else {
		return fmt.Errorf("unsupported project structure")
	}
}

func uploadMTA(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	if options.GroupID == "" {
		return fmt.Errorf("the 'groupId' parameter needs to be provided for MTA projects")
	}
	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository)
	var mtaPath string
	if err == nil {
		exists, _ := utils.fileExists("mta.yaml")
		if exists {
			mtaPath = "mta.yaml"
			// Give this file precedence, but it would be even better if
			// ProjectStructure could be asked for the mta file it detected.
		} else {
			// This will fail anyway if the file doesn't exist
			mtaPath = "mta.yml"
		}
		err = setVersionFromMtaFile(utils, uploader, mtaPath)
	}
	var artifactID = options.ArtifactID
	if artifactID == "" {
		artifactID = utils.getEnvParameter(".pipeline/commonPipelineEnvironment/configuration", "artifactId")
		if artifactID == "" {
			err = fmt.Errorf("the 'artifactId' parameter was not provided and could not be retrieved from the Common Pipeline Environment")
		} else {
			log.Entry().Debugf("mtar artifact id from CPE: '%s'", artifactID)
		}
	}
	if err == nil {
		err = addArtifact(utils, uploader, mtaPath, "", "yaml", artifactID)
	}
	if err == nil {
		mtarFilePath := utils.getEnvParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		log.Entry().Debugf("mtar file path: '%s'", mtarFilePath)
		err = addArtifact(utils, uploader, mtarFilePath, "", "mtar", artifactID)
	}
	if err == nil {
		err = uploadArtifacts(uploader, options)
	}
	return err
}

type mtaYaml struct {
	ID      string `json:"ID"`
	Version string `json:"version"`
}

func setVersionFromMtaFile(utils nexusUploadUtils, uploader nexus.Uploader, filePath string) error {
	mtaYamlContent, err := utils.fileRead(filePath)
	if err != nil {
		return err
	}
	return setVersionFromMtaYaml(uploader, mtaYamlContent)
}

func setVersionFromMtaYaml(uploader nexus.Uploader, mtaYamlContent []byte) error {
	var mtaYaml mtaYaml
	err := yaml.Unmarshal(mtaYamlContent, &mtaYaml)
	if err != nil {
		return err
	}
	return uploader.SetArtifactsVersion(mtaYaml.Version)
}

func uploadArtifacts(uploader nexus.Uploader, options *nexusUploadOptions) error {
	artifacts := uploader.GetArtifacts()
	if len(artifacts) == 0 {
		return errors.New("no artifacts to upload")
	}

	execRunner := command.Command{}
	execRunner.Stdout(ioutil.Discard)
	execRunner.Stderr(ioutil.Discard)

	var defines []string
	defines = append(defines, "-Durl=http://"+uploader.GetBaseURL())

	file := ""
	files := ""
	classifiers := ""
	types := ""
	artifactId := artifacts[0].ID

	for i, artifact := range artifacts {
		if i == 0 {
			file = artifact.File
		} else {
			if i > 1 {
				files += ","
				classifiers += ","
				types += ","
			}
			files += artifact.File
			classifiers += artifact.Classifier
			types += artifact.Type
		}
		if artifactId != artifact.ID {
			return fmt.Errorf(
				"cannot deploy artifacts with different IDs in one run (%s vs. %s)",
				artifactId, artifact.ID)
		}
	}

	defines = append(defines, "-DgroupId="+options.GroupID)
	defines = append(defines, "-DartifactId="+artifactId)
	defines = append(defines, "-Dversion="+uploader.GetArtifactsVersion())
	defines = append(defines, "-Dfile="+file)
	defines = append(defines, "-DgeneratePom=false")
	if len(files) > 0 {
		defines = append(defines, "-Dfiles="+files)
		defines = append(defines, "-Dclassifiers="+classifiers)
		defines = append(defines, "-Dtypes="+types)
	}

	mavenOptions := maven.ExecuteOptions{
		Goals:        []string{"deploy:deploy-file"},
		Defines:      defines,
		ReturnStdout: false,
	}
	_, err := maven.Execute(&mavenOptions, &execRunner)
	if err != nil {
		return err
	}
	return nil
}
func uploadMaven(uploader nexus.Uploader, options *nexusUploadOptions) error {
	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository)
	if err != nil {
		return err
	}

	execRunner := command.Command{}
	execRunner.Stdout(ioutil.Discard)
	execRunner.Stderr(ioutil.Discard)

	// This is the ID which maven will look up in the local settings file to find login credentials
	// TODO: Create a temporary settings file, store credential keys so they can be substituted from env variables
	repositoryId := options.Repository

	altRepository := repositoryId + "::default::http://" + uploader.GetBaseURL()

	var defines []string
	defines = append(defines, "-Dmaven.test.skip")
	defines = append(defines, "-DaltDeploymentRepository="+altRepository)

	testModulesExcludes := maven.GetTestModulesExcludes()
	if testModulesExcludes != nil {
		defines = append(defines, testModulesExcludes...)
	}

	mavenOptions := maven.ExecuteOptions{
		Goals:        []string{"deploy"},
		Defines:      defines,
		ReturnStdout: false,
	}
	_, err = maven.Execute(&mavenOptions, &execRunner)
	if err != nil {
		return err
	}
	return nil
}

func addArtifact(utils nexusUploadUtils, uploader nexus.Uploader, filePath, classifier, fileType, id string) error {
	exists, _ := utils.fileExists(filePath)
	if !exists {
		return fmt.Errorf("artifact file not found '%s'", filePath)
	}
	artifact := nexus.ArtifactDescription{
		File:       filePath,
		Type:       fileType,
		Classifier: classifier,
		ID:         id,
	}
	return uploader.AddArtifact(artifact)
}
