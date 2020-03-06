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

type projectStructure interface {
	UsesMta() bool
	UsesMaven() bool
}

type mavenExecutor struct {
	execRunner command.Command
}

type mavenEvaluator interface {
	evaluateProperty(pomFile, expression string) (string, error)
}

func nexusUpload(options nexusUploadOptions, telemetryData *telemetry.CustomData) {
	uploader := nexus.Upload{Username: options.User, Password: options.Password}
	projectStructure := piperutils.ProjectStructure{}
	fileUtils := piperutils.Files{}
	evaluator := mavenExecutor{execRunner: command.Command{}}

	err := runNexusUpload(&options, &uploader, &projectStructure, &fileUtils, &evaluator)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNexusUpload(options *nexusUploadOptions, uploader nexus.Uploader, projectStructure projectStructure,
	fileUtils piperutils.FileUtils, evaluator mavenEvaluator) error {

	if projectStructure.UsesMta() {
		log.Entry().Info("MTA project structure detected")
		return uploadMTA(uploader, fileUtils, options)
	} else if projectStructure.UsesMaven() {
		log.Entry().Info("Maven project structure detected")
		return uploadMaven(uploader, fileUtils, evaluator, options)
	} else {
		return fmt.Errorf("unsupported project structure")
	}
}

func uploadMTA(uploader nexus.Uploader, fileUtils piperutils.FileUtils, options *nexusUploadOptions) error {
	if options.GroupID == "" {
		return fmt.Errorf("the 'groupID' parameter needs to be provided for MTA projects")
	}
	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository, options.GroupID)
	if err == nil {
		exists, _ := fileUtils.FileExists("mta.yaml")
		if exists {
			// Give this file precedence, but it would be even better if
			// ProjectStructure could be asked for the mta file it detected.
			err = setVersionFromMtaFile(uploader, fileUtils, "mta.yaml")
		} else {
			// This will fail anyway if the file doesn't exist
			err = setVersionFromMtaFile(uploader, fileUtils, "mta.yml")
		}
	}
	if err == nil {
		artifactID := options.ArtifactID
		if artifactID == "" {
			artifactID = piperenv.GetParameter(".pipeline/commonPipelineEnvironment/configuration", "artifactId")
			log.Entry().Debugf("mtar artifact id from CPE: '%s'", artifactID)
		}
		err = uploader.AddArtifact(nexus.ArtifactDescription{File: "mta.yaml", Type: "yaml", Classifier: "", ID: options.ArtifactID})
	}
	if err == nil {
		mtarFilePath := piperenv.GetParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		log.Entry().Debugf("mtar file path: '%s'", mtarFilePath)
		err = uploader.AddArtifact(nexus.ArtifactDescription{File: mtarFilePath, Type: "mtar", Classifier: "", ID: options.ArtifactID})
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

func setVersionFromMtaFile(uploader nexus.Uploader, fileUtils piperutils.FileUtils, filePath string) error {
	mtaYamlContent, err := fileUtils.FileRead(filePath)
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

var errPomNotFound error = errors.New("pom.xml not found")

func uploadMaven(uploader nexus.Uploader, fileUtils piperutils.FileUtils, evaluator mavenEvaluator,
	options *nexusUploadOptions) error {
	err := uploadMavenArtifacts(uploader, fileUtils, evaluator, options,
		"", "target", "")
	if err != nil {
		return err
	}

	// Test if a sub-folder "application" exists and upload the artifacts from there as well.
	// This means there are built-in assumptions about the project structure (archetype),
	// that nexusUpload supports. To make this more flexible should be the scope of another PR.
	err = uploadMavenArtifacts(uploader, fileUtils, evaluator, options,
		"application", "application/target", options.AdditionalClassifiers)
	if err == errPomNotFound {
		// Ignore for missing application module
		return nil
	}
	return err
}

func uploadMavenArtifacts(uploader nexus.Uploader, fileUtils piperutils.FileUtils, evaluator mavenEvaluator,
	options *nexusUploadOptions, pomPath, targetFolder, additionalClassifiers string) error {
	var err error

	pomFile := composeFilePath(pomPath, "pom", "xml")
	exists, _ := fileUtils.FileExists(pomFile)
	if !exists {
		return errPomNotFound
	}
	groupID, err := evaluator.evaluateProperty(pomFile, "project.groupId")
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
		artifactID, err = evaluator.evaluateProperty(pomFile, "project.artifactId")
	}
	var artifactsVersion string
	if err == nil {
		artifactsVersion, err = evaluator.evaluateProperty(pomFile, "project.version")
	}
	if err == nil {
		err = uploader.SetArtifactsVersion(artifactsVersion)
	}
	if err == nil {
		artifact := nexus.ArtifactDescription{
			File:       pomFile,
			Type:       "pom",
			Classifier: "",
			ID:         artifactID,
		}
		err = uploader.AddArtifact(artifact)
	}
	if err == nil {
		err = addTargetArtifact(pomFile, targetFolder, artifactID, uploader, evaluator)
	}
	if err == nil {
		err = addAdditionalClassifierArtifacts(additionalClassifiers, targetFolder, artifactID, uploader)
	}
	if err == nil {
		err = uploader.UploadArtifacts()
	}
	return err
}

func addTargetArtifact(pomFile, targetFolder, artifactID string, uploader nexus.Uploader, evaluator mavenEvaluator) error {
	packaging, err := evaluator.evaluateProperty(pomFile, "project.packaging")
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
	finalName, err := evaluator.evaluateProperty(pomFile, "project.build.finalName")
	if err != nil || finalName == "" {
		// NOTE: The error should be ignored, and the finalName built as Maven would from artifactId and so on.
		// But it seems this expression always resolves, even if finalName is nowhere declared in the pom.xml
		return err
	}
	filePath := composeFilePath(targetFolder, finalName, packaging)
	artifact := nexus.ArtifactDescription{
		File:       filePath,
		Type:       packaging,
		Classifier: "",
		ID:         artifactID,
	}
	return uploader.AddArtifact(artifact)
}

func addAdditionalClassifierArtifacts(additionalClassifiers, targetFolder, artifactID string, uploader nexus.Uploader) error {
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
		artifact := nexus.ArtifactDescription{
			File:       filePath,
			Type:       classifier.FileType,
			Classifier: classifier.Classifier,
			ID:         artifactID,
		}
		err = uploader.AddArtifact(artifact)
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

func (m *mavenExecutor) evaluateProperty(pomFile, expression string) (string, error) {
	m.execRunner.Stdout(ioutil.Discard)
	m.execRunner.Stderr(ioutil.Discard)

	expressionDefine := "-Dexpression=" + expression

	options := maven.ExecuteOptions{
		PomPath:      pomFile,
		Goals:        []string{"org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"},
		Defines:      []string{expressionDefine, "-DforceStdout", "-q"},
		ReturnStdout: true,
	}
	value, err := maven.Execute(&options, &m.execRunner)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(value, "null object or invalid expression") {
		return "", fmt.Errorf("expression '%s' in file '%s' could not be resolved", expression, pomFile)
	}
	log.Entry().Debugf("Evaluated expression '%s' in file '%s' as '%s'\n", expression, pomFile, value)
	return value, nil
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
