package cmd

import (
	"encoding/json"
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

func nexusUpload(config nexusUploadOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	runNexusUpload(&config, telemetryData, &c)
}

func runNexusUpload(config *nexusUploadOptions, telemetryData *telemetry.CustomData, command execRunner) {
	projectStructure := piperutils.ProjectStructure{}

	nexusClient := nexus.Upload{Username: config.User, Password: config.Password}

	if projectStructure.UsesMta() {
		//		if GeneralConfig.Verbose {
		log.Entry().Info("MTA project structure detected")
		//		}
		uploadMTA(&nexusClient, config)
	} else if projectStructure.UsesMaven() {
		//		if GeneralConfig.Verbose {
		log.Entry().Info("Maven project structure detected")
		//		}
		uploadMaven(&nexusClient, config)
	} else {
		log.Entry().Fatal("Unsupported project structure")
	}
}

func uploadMTA(nexusClient *nexus.Upload, config *nexusUploadOptions) {
	if config.GroupID == "" {
		log.Entry().Fatal("The 'groupID' parameter needs to be provided for MTA projects")
	}
	err := nexusClient.SetBaseURL(config.Url, config.Version, config.Repository, config.GroupID)
	if err == nil {
		err = setVersionFromMtaYaml(nexusClient)
	}
	if err == nil {
		artifactID := config.ArtifactID
		if artifactID == "" {
			artifactID = piperenv.GetParameter(".pipeline/commonPipelineEnvironment/configuration", "artifactId")
		}
		// TODO: Read artifactID from commonPipelineEnvironment if not given via config
		// commonPipelineEnvironment.configuration.artifactId
		err = nexusClient.AddArtifact(nexus.ArtifactDescription{File: "mta.yaml", Type: "yaml", Classifier: "", ID: config.ArtifactID})
	}
	if err == nil {
		//TODO: do proper way to find name/path of mta file
		mtarFilePath := piperenv.GetParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		fmt.Println(mtarFilePath)
		err = nexusClient.AddArtifact(nexus.ArtifactDescription{File: mtarFilePath, Type: "mtar", Classifier: "", ID: config.ArtifactID})
	}
	if err != nil {
		log.Entry().WithError(err).Fatal()
	}

	nexusClient.UploadArtifacts()
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

func uploadMaven(nexusClient *nexus.Upload, config *nexusUploadOptions) {
	uploadMavenArtifacts(nexusClient, config, "", "target", "")
	uploadMavenArtifacts(nexusClient, config, "application", "application/target",
		config.AdditionalClassifiers)
}

func uploadMavenArtifacts(nexusClient *nexus.Upload, config *nexusUploadOptions, pomPath, targetFolder, additionalClassifiers string) {
	var err error

	pomFile := composeFilePath(pomPath, "pom", "xml")
	groupID, err := evaluateMavenProperty(pomFile, "project.groupId")
	if groupID == "" {
		groupID = config.GroupID
	}
	if err == nil {
		err = nexusClient.SetBaseURL(config.Url, config.Version, config.Repository, groupID)
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

	if err != nil {
		log.Entry().WithError(err).Fatal()
	}

	nexusClient.UploadArtifacts()
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
			return errors.New(
				fmt.Sprintf("Invalid additional classifier description (classifier: '%s', type: '%s')",
					classifier.Classifier, classifier.FileType))
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
		return "", errors.New(
			fmt.Sprintf("Expression could not be resolved, property not found or invalid expression '%s'",
				expression))
	}
	//	if GeneralConfig.Verbose {
	log.Entry().Infof("Evaluated expression '%s' in file '%s' as '%s'\n", expression, pomFile, value)
	//	}
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
