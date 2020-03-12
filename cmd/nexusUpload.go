package cmd

import (
	"fmt"
	"github.com/pkg/errors"

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
	getExecRunner() execRunner
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

func (u *utilsBundle) getExecRunner() execRunner {
	execRunner := command.Command{}
	execRunner.Stdout(log.Entry().Writer())
	execRunner.Stderr(log.Entry().Writer())
	return &execRunner
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
		return uploadMaven(utils, uploader, options)
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
		err = uploadArtifactsMTA(utils, uploader, options)
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
		return fmt.Errorf("could not read from required project descriptor file '%s'",
			filePath)
	}
	return setVersionFromMtaYaml(uploader, mtaYamlContent, filePath)
}

func setVersionFromMtaYaml(uploader nexus.Uploader, mtaYamlContent []byte, filePath string) error {
	var mtaYaml mtaYaml
	err := yaml.Unmarshal(mtaYamlContent, &mtaYaml)
	if err != nil {
		// Eat the original error as it is unhelpful and confusingly mentions JSON, while the
		// user thinks it should parse YAML (it is transposed by the implementation).
		return fmt.Errorf("failed to parse contents of the project descriptor file '%s'",
			filePath)
	}
	err = uploader.SetArtifactsVersion(mtaYaml.Version)
	if err != nil {
		return fmt.Errorf("the project descriptor file '%s' has an invalid version: %w",
			filePath, err)
	}
	return nil
}

func uploadArtifactsMTA(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	artifacts := uploader.GetArtifacts()
	if len(artifacts) == 0 {
		return errors.New("no artifacts to upload")
	}

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
	_, err := maven.Execute(&mavenOptions, utils.getExecRunner())
	if err != nil {
		return fmt.Errorf("uploading artifacts failed: %w", err)
	}
	return nil
}

func uploadMaven(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	if pomExists, _ := utils.fileExists("pom.xml"); !pomExists {
		return errors.New("pom.xml not found")
	}

	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository)
	if err != nil {
		return err
	}

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
	_, err = maven.Execute(&mavenOptions, utils.getExecRunner())
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
