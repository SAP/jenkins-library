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
	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	runNexusUpload(&options, telemetryData)
}

func runNexusUpload(options *nexusUploadOptions, telemetryData *telemetry.CustomData) {
	projectStructure := piperutils.ProjectStructure{}

	nexusClient := nexus.Upload{Username: options.User, Password: options.Password}

	if projectStructure.UsesMta() {
		//		if GeneralConfig.Verbose {
		log.Entry().Info("MTA project structure detected")
		//		}
		uploadMTA(&nexusClient, options)
	} else if projectStructure.UsesMaven() {
		//		if GeneralConfig.Verbose {
		log.Entry().Info("Maven project structure detected")
		//		}
		uploadMaven(&nexusClient, options)
	} else {
		log.Entry().Fatal("Unsupported project structure")
	}
}

func uploadMTA(nexusClient *nexus.Upload, options *nexusUploadOptions) {
	if options.GroupID == "" {
		log.Entry().Fatal("The 'groupID' parameter needs to be provided for MTA projects")
	}
	err := nexusClient.SetBaseURL(options.Url, options.Version, options.Repository, options.GroupID)
	if err == nil {
		err = setVersionFromMtaYaml(nexusClient)
	}
	if err == nil {
		artifactID := options.ArtifactID
		if artifactID == "" {
			artifactID = piperenv.GetParameter(".pipeline/commonPipelineEnvironment/optionsuration", "artifactId")
		}
		// TODO: Read artifactID from commonPipelineEnvironment if not given via options
		// commonPipelineEnvironment.optionsuration.artifactId
		err = nexusClient.AddArtifact(nexus.ArtifactDescription{File: "mta.yaml", Type: "yaml", Classifier: "", ID: options.ArtifactID})
	}
	if err == nil {
		//TODO: do proper way to find name/path of mta file
		mtarFilePath := piperenv.GetParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		fmt.Println(mtarFilePath)
		err = nexusClient.AddArtifact(nexus.ArtifactDescription{File: mtarFilePath, Type: "mtar", Classifier: "", ID: options.ArtifactID})
	}
	if err == nil {
		err = nexusClient.UploadArtifacts()
	}
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

type mtaYaml struct {
	ID      string `json:"ID"`
	Version string `json:"version"`
}

func setVersionFromMtaYaml(nexusClient *nexus.Upload) error {
	var mtaYaml mtaYaml
	mtaYamContent, err := ioutil.ReadFile("mta.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(mtaYamContent, &mtaYaml)
	if err != nil {
		return err
	}
	return nexusClient.SetArtifactsVersion(mtaYaml.Version)
}

var errPomNotFound error = errors.New("pom.xml not found")

func uploadMaven(nexusClient *nexus.Upload, options *nexusUploadOptions) {
	err := uploadMavenArtifacts(nexusClient, options, "", "target", "")
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	// Test if a sub-folder "application" exists and upload the artifacts from there as well.
	// This means there are built-in assumptions about the project structure (archetype),
	// that nexusUpload supports. To make this more flexible should be the scope of another PR.
	err = uploadMavenArtifacts(nexusClient, options, "application", "application/target", options.AdditionalClassifiers)
	if err == errPomNotFound {
		// Ignore
	} else if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func uploadMavenArtifacts(nexusClient *nexus.Upload, options *nexusUploadOptions, pomPath, targetFolder, additionalClassifiers string) error {
	var err error

	pomFile := composeFilePath(pomPath, "pom", "xml")
	stat, err := os.Stat(pomFile)
	if err != nil || stat.IsDir() {
		return errPomNotFound
	}

	// Begin testing effective POM generation
	effectivePomFile := composeFilePath(pomPath, "effectivePom", "xml")
	m2Path := "s4hana_pipeline/maven_local_repo"
	err = generateEffectivePOM(pomFile, effectivePomFile, m2Path, nil)
	if err != nil {
		return fmt.Errorf("failed to generate effective POM: %w", err)
	}
	pomFile = effectivePomFile
	// End testing effective POM generation

	groupID, err := evaluateMavenProperty(pomFile, "project.groupId")
	if groupID == "" {
		groupID = options.GroupID
		// Reset error
		err = nil
	}
	if err == nil {
		err = nexusClient.SetBaseURL(options.Url, options.Version, options.Repository, groupID)
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
		err = nexusClient.SetArtifactsVersion(artifactsVersion)
	}
	if err == nil {
		artifact := nexus.ArtifactDescription{
			File:       pomFile,
			Type:       "pom",
			Classifier: "",
			ID:         artifactID,
		}
		err = nexusClient.AddArtifact(artifact)
	}
	if err == nil {
		err = addTargetArtifact(pomFile, targetFolder, artifactID, nexusClient)
	}
	if err == nil {
		err = addAdditionalClassifierArtifacts(additionalClassifiers, targetFolder, artifactID, nexusClient)
	}
	if err == nil {
		err = nexusClient.UploadArtifacts()
	}
	return err
}

func addTargetArtifact(pomFile, targetFolder, artifactID string, nexusClient *nexus.Upload) error {
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
	return nexusClient.AddArtifact(artifact)
}

func addAdditionalClassifierArtifacts(additionalClassifiers, targetFolder, artifactID string, nexusClient *nexus.Upload) error {
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
		err = nexusClient.AddArtifact(artifact)
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
	//	if GeneralConfig.Verbose {
	log.Entry().Infof("Evaluated expression '%s' in file '%s' as '%s'\n", expression, pomFile, value)
	//	}
	return value, nil
}

func generateEffectivePOM(pomFile, effectivePomFile, m2Path string, execRunner *command.Command) error {
	if execRunner == nil {
		execRunner = &command.Command{}
		execRunner.Stdout(ioutil.Discard)
		execRunner.Stderr(ioutil.Discard)
	}

	options := maven.ExecuteOptions{
		PomPath:      pomFile,
		M2Path:       m2Path,
		Goals:        []string{"help:effective-pom"},
		Defines:      []string{"-Doutput="+effectivePomFile},
	}
	_, err := maven.Execute(&options, execRunner)
	return err
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
