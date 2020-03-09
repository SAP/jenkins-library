package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
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

func nexusUpload(options nexusUploadOptions, telemetryData *telemetry.CustomData) {
	utils := newUtilsBundle()
	uploader := nexus.Upload{Username: options.User, Password: options.Password}

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
		return fmt.Errorf("the 'groupID' parameter needs to be provided for MTA projects")
	}
	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository, options.GroupID)
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
	if err == nil {
		if artifactID == "" {
			artifactID = utils.getEnvParameter(".pipeline/commonPipelineEnvironment/configuration", "artifactId")
			log.Entry().Debugf("mtar artifact id from CPE: '%s'", artifactID)
		}
		err = addArtifact(utils, uploader, mtaPath, "", "yaml", artifactID)
	}
	if err == nil {
		mtarFilePath := utils.getEnvParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		log.Entry().Debugf("mtar file path: '%s'", mtarFilePath)
		err = addArtifact(utils, uploader, mtarFilePath, "", "mtar", artifactID)
	}
	if err == nil {
		err = uploader.UploadArtifacts()
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

var errPomNotFound = errors.New("pom.xml not found")

func uploadMaven(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	err := uploadMavenArtifacts(utils, uploader, options, "", "target", "")
	if err != nil {
		return err
	}

	// Test if a sub-folder "application" exists and upload the artifacts from there as well.
	// This means there are built-in assumptions about the project structure (archetype),
	// that nexusUpload supports. To make this more flexible should be the scope of another PR.
	err = uploadMavenArtifacts(utils, uploader, options, "application", "application/target",
		options.AdditionalClassifiers)
	if err == errPomNotFound {
		// Ignore for missing application module
		return nil
	}
	return err
}

func uploadMavenArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions,
	pomPath, targetFolder, additionalClassifiers string) error {
	var err error

	pomFile := composeFilePath(pomPath, "pom", "xml")
	exists, _ := utils.fileExists(pomFile)
	if !exists {
		return errPomNotFound
	}
	groupID, err := utils.evaluateProperty(pomFile, "project.groupId")
	if groupID == "" {
		groupID = options.GroupID
		// Reset error
		err = nil
	}
	if err == nil {
		err = uploader.SetBaseURL(options.Url, options.Version, options.Repository, groupID)
	}
	var artifactID string
	if err == nil {
		artifactID, err = utils.evaluateProperty(pomFile, "project.artifactId")
	}
	var artifactsVersion string
	if err == nil {
		artifactsVersion, err = utils.evaluateProperty(pomFile, "project.version")
	}
	if err == nil {
		err = uploader.SetArtifactsVersion(artifactsVersion)
	}
	if err == nil {
		err = addArtifact(utils, uploader, pomFile, "", "pom", artifactID)
	}
	if err == nil {
		err = addTargetArtifact(utils, uploader, pomFile, targetFolder, artifactID)
	}
	if err == nil {
		err = addAdditionalClassifierArtifacts(utils, uploader, additionalClassifiers, targetFolder, artifactID)
	}
	if err == nil {
		err = uploader.UploadArtifacts()
	}
	return err
}

func addTargetArtifact(utils nexusUploadUtils, uploader nexus.Uploader, pomFile, targetFolder, artifactID string) error {
	packaging, err := utils.evaluateProperty(pomFile, "project.packaging")
	if err != nil {
		return err
	}
	if packaging == "pom" {
		// Only pom.xml itself is the artifact
		return nil
	}
	if packaging == "" {
		packaging = "jar"
	}
	finalName, err := utils.evaluateProperty(pomFile, "project.build.finalName")
	if err != nil || finalName == "" {
		// NOTE: The error should be ignored, and the finalName built as Maven would from artifactId and so on.
		// But it seems this expression always resolves, even if finalName is nowhere declared in the pom.xml
		return err
	}
	filePath := composeFilePath(targetFolder, finalName, packaging)
	return addArtifact(utils, uploader, filePath, "", packaging, artifactID)
}

func addAdditionalClassifierArtifacts(utils nexusUploadUtils, uploader nexus.Uploader,
	additionalClassifiers, targetFolder, artifactID string) error {
	if additionalClassifiers == "" {
		return nil
	}
	classifiers, err := getClassifiers(additionalClassifiers)
	if err != nil {
		return err
	}
	for _, classifier := range classifiers {
		if classifier.Classifier == "" || classifier.FileType == "" {
			return fmt.Errorf("invalid additional classifier description (classifier: '%s', type: '%s')",
				classifier.Classifier, classifier.FileType)
		}
		filePath := composeFilePath(targetFolder, artifactID+"-"+classifier.Classifier, classifier.FileType)
		err = addArtifact(utils, uploader, filePath, classifier.FileType, classifier.Classifier, artifactID)
		if err != nil {
			return err
		}
	}
	return nil
}

func composeFilePath(folder, name, extension string) string {
	fileName := name + "." + extension
	return filepath.Join(folder, fileName)
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

type classifierDescription struct {
	Classifier string `json:"classifier"`
	FileType   string `json:"type"`
}

func getClassifiers(classifiersAsJSON string) ([]classifierDescription, error) {
	var classifiers []classifierDescription
	err := json.Unmarshal([]byte(classifiersAsJSON), &classifiers)
	return classifiers, err
}
