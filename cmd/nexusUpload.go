package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
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

func nexusUpload(options nexusUploadOptions, telemetryData *telemetry.CustomData) {
	uploader := nexus.Upload{Username: options.User, Password: options.Password}

	err := runNexusUpload(&options, &uploader, telemetryData)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNexusUpload(options *nexusUploadOptions, uploader nexus.Uploader, telemetryData *telemetry.CustomData) error {
	projectStructure := piperutils.ProjectStructure{}

	if projectStructure.UsesMta() {
		log.Entry().Info("MTA project structure detected")
		return uploadMTA(uploader, options)
	} else if projectStructure.UsesMaven() {
		log.Entry().Info("Maven project structure detected")
		return uploadMaven(uploader, options)
	} else {
		return fmt.Errorf("unsupported project structure")
	}
}

func uploadMTA(uploader nexus.Uploader, options *nexusUploadOptions) error {
	if options.GroupID == "" {
		return fmt.Errorf("the 'groupID' parameter needs to be provided for MTA projects")
	}
	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository, options.GroupID)
	if err == nil {
		exists, _ := piperutils.FileExists("mta.yaml")
		if exists {
			// Give this file precedence, but it would be even better if
			// ProjectStructure could be asked for the mta file it detected.
			err = setVersionFromMtaFile(uploader, "mta.yaml")
		} else {
			// This will fail anyway if the file doesn't exist
			err = setVersionFromMtaFile(uploader, "mta.yml")
		}
	}
	if err == nil {
		artifactID := options.ArtifactID
		if artifactID == "" {
			artifactID = piperenv.GetParameter(".pipeline/commonPipelineEnvironment/configuration", "artifactId")
		}
		err = uploader.AddArtifact(nexus.ArtifactDescription{File: "mta.yaml", Type: "yaml", Classifier: "", ID: options.ArtifactID})
	}
	if err == nil {
		mtarFilePath := piperenv.GetParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		fmt.Println(mtarFilePath)
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

func setVersionFromMtaFile(uploader nexus.Uploader, filePath string) error {
	mtaYamlContent, err := ioutil.ReadFile(filePath)
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

func uploadMaven(uploader nexus.Uploader, options *nexusUploadOptions) error {
	err := uploadMavenArtifacts(uploader, options, "", "target", "")
	if err != nil {
		return err
	}

	// Test if a sub-folder "application" exists and upload the artifacts from there as well.
	// This means there are built-in assumptions about the project structure (archetype),
	// that nexusUpload supports. To make this more flexible should be the scope of another PR.
	err = uploadMavenArtifacts(uploader, options, "application", "application/target", options.AdditionalClassifiers)
	if err == errPomNotFound {
		// Ignore for missing application module
		return nil
	}
	return err
}

func uploadMavenArtifacts(uploader nexus.Uploader, options *nexusUploadOptions, pomPath, targetFolder, additionalClassifiers string) error {
	var err error

	pomFile := composeFilePath(pomPath, "pom", "xml")
	stat, err := os.Stat(pomFile)
	if err != nil || stat.IsDir() {
		return errPomNotFound
	}
	groupID, err := evaluateMavenProperty(pomFile, "project.groupId")
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
		artifactID, err = evaluateMavenProperty(pomFile, "project.artifactId")
	}
	var artifactsVersion string
	if err == nil {
		artifactsVersion, err = evaluateMavenProperty(pomFile, "project.version")
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
		err = addTargetArtifact(pomFile, targetFolder, artifactID, uploader)
	}
	if err == nil {
		err = addAdditionalClassifierArtifacts(additionalClassifiers, targetFolder, artifactID, uploader)
	}
	if err == nil {
		err = uploader.UploadArtifacts()
	}
	return err
}

func addTargetArtifact(pomFile, targetFolder, artifactID string, uploader nexus.Uploader) error {
	packaging, err := evaluateMavenProperty(pomFile, "project.packaging")
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
	finalName, err := evaluateMavenProperty(pomFile, "project.build.finalName")
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
	path := name + "." + extension
	if folder != "" {
		path = folder + "/" + path
	}
	return path
}

func evaluateMavenProperty(pomFile, expression string) (string, error) {
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
	if GeneralConfig.Verbose {
		log.Entry().Infof("Evaluated expression '%s' in file '%s' as '%s'\n", expression, pomFile, value)
	}
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
