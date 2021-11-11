package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

type mavenBuildSettingsInfo struct {
	Profiles                    []string `json:"profiles,omitempty"`
	Publish                     bool     `json:"publish,omitempty"`
	CreateBOM                   bool     `json:"createBOM,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
}

type buildSettings struct {
	MavenBuild []mavenBuildSettingsInfo `json:mavenBuild`
}

func mavenBuild(config mavenBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *mavenBuildCommonPipelineEnvironment) {
	utils := maven.NewUtilsBundle()

	err := runMavenBuild(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenBuild(config *mavenBuildOptions, telemetryData *telemetry.CustomData, utils maven.Utils, commonPipelineEnvironment *mavenBuildCommonPipelineEnvironment) error {

	var flags = []string{"-update-snapshots", "--batch-mode"}

	if len(config.Profiles) > 0 {
		flags = append(flags, "--activate-profiles", strings.Join(config.Profiles, ","))
	}

	exists, _ := utils.FileExists("integration-tests/pom.xml")
	if exists {
		flags = append(flags, "-pl", "!integration-tests")
	}

	var defines []string
	var goals []string

	goals = append(goals, "org.jacoco:jacoco-maven-plugin:prepare-agent")

	if config.Flatten {
		goals = append(goals, "flatten:flatten")
		defines = append(defines, "-Dflatten.mode=resolveCiFriendliesOnly", "-DupdatePomFile=true")
	}

	if config.CreateBOM {
		goals = append(goals, "org.cyclonedx:cyclonedx-maven-plugin:makeAggregateBom")
		createBOMConfig := []string{
			"-DschemaVersion=1.2",
			"-DincludeBomSerialNumber=true",
			"-DincludeCompileScope=true",
			"-DincludeProvidedScope=true",
			"-DincludeRuntimeScope=true",
			"-DincludeSystemScope=true",
			"-DincludeTestScope=false",
			"-DincludeLicenseText=false",
			"-DoutputFormat=xml",
		}
		defines = append(defines, createBOMConfig...)
	}

	if config.Verify {
		goals = append(goals, "verify")
	} else {
		goals = append(goals, "install")
	}

	mavenOptions := maven.ExecuteOptions{
		Flags:                       flags,
		Goals:                       goals,
		Defines:                     defines,
		PomPath:                     config.PomPath,
		ProjectSettingsFile:         config.ProjectSettingsFile,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		M2Path:                      config.M2Path,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
	}

	_, err := maven.Execute(&mavenOptions, utils)

	if err == nil {
		if config.Publish && !config.Verify {
			log.Entry().Infof("publish detected, running mvn deploy")

			if (len(config.AltDeploymentRepositoryID) > 0) && (len(config.AltDeploymentRepositoryPassword) > 0) && (len(config.AltDeploymentRepositoryUser) > 0) {
				projectSettingsFilePath, err := createOrUpdateProjectSettingsXML(config.ProjectSettingsFile, config.AltDeploymentRepositoryID, config.AltDeploymentRepositoryUser, config.AltDeploymentRepositoryPassword, utils)
				if err != nil {
					return errors.Wrap(err, "Could not create or update project settings xml")
				}
				mavenOptions.ProjectSettingsFile = projectSettingsFilePath
			}

			deployFlags := []string{"-Dmaven.main.skip=true", "-Dmaven.test.skip=true", "-Dmaven.install.skip=true"}
			if (len(config.AltDeploymentRepositoryID) > 0) && (len(config.AltDeploymentRepositoryURL) > 0) {
				deployFlags = append(deployFlags, "-DaltDeploymentRepository="+config.AltDeploymentRepositoryID+"::default::"+config.AltDeploymentRepositoryURL)
			}

			downloadClient := &piperhttp.Client{}
			runner := &command.Command{}
			fileUtils := &piperutils.Files{}
			if len(config.CustomTLSCertificateLinks) > 0 {
				if err := loadRemoteRepoCertificates(config.CustomTLSCertificateLinks, downloadClient, &deployFlags, runner, fileUtils, config.JavaCaCertFilePath); err != nil {
					log.SetErrorCategory(log.ErrorInfrastructure)
					return err
				}
			}

			mavenOptions.Flags = deployFlags
			mavenOptions.Goals = []string{"deploy"}
			mavenOptions.Defines = []string{}
			_, err := maven.Execute(&mavenOptions, utils)
			return err
		} else {
			log.Entry().Infof("publish not detected, ignoring maven deploy")
		}
	}

	builSettingsErr := createMavenBuildSettingsInfo(config, commonPipelineEnvironment)

	if builSettingsErr != nil {
		log.Entry().Warnf("failed to create build settings info : ''%v", builSettingsErr)
	}

	return err
}

func createMavenBuildSettingsInfo(config *mavenBuildOptions, commonPipelineEnvironment *mavenBuildCommonPipelineEnvironment) error {
	currentBuildSettingsInfo := mavenBuildSettingsInfo{
		CreateBOM:                   config.CreateBOM,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		Profiles:                    config.Profiles,
		Publish:                     config.Publish,
	}
	var jsonMap map[string][]interface{}

	if len(config.BuildSettingsInfo) > 0 {

		err := json.Unmarshal([]byte(config.BuildSettingsInfo), &jsonMap)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal existing build settings json '%v'", config.BuildSettingsInfo)
		}

		if mavenBuild, exist := jsonMap["mavenBuild"]; exist {
			if reflect.TypeOf(mavenBuild).Kind() == reflect.Slice {
				jsonMap["mavenBuild"] = append(mavenBuild, currentBuildSettingsInfo)
			}
		} else {
			var settings []interface{}
			settings = append(settings, currentBuildSettingsInfo)
			jsonMap["mavenBuild"] = settings
		}

		newJsonMap, err := json.Marshal(&jsonMap)

		if err != nil {
			return errors.Wrapf(err, "Creating build settings failed with json marshalling")
		}

		commonPipelineEnvironment.custom.buildSettingsInfo = string(newJsonMap)

	} else {
		var settings []mavenBuildSettingsInfo
		settings = append(settings, currentBuildSettingsInfo)
		jsonResult, err := json.Marshal(buildSettings{
			MavenBuild: settings,
		})
		if err != nil {
			return errors.Wrapf(err, "Creating build settings failed with json marshalling")
		} else {
			commonPipelineEnvironment.custom.buildSettingsInfo = string(jsonResult)
		}
	}

	return nil

}

func createOrUpdateProjectSettingsXML(projectSettingsFile string, altDeploymentRepositoryID string, altDeploymentRepositoryUser string, altDeploymentRepositoryPassword string, utils maven.Utils) (string, error) {
	if len(projectSettingsFile) > 0 {
		projectSettingsFilePath, err := maven.UpdateProjectSettingsXML(projectSettingsFile, altDeploymentRepositoryID, altDeploymentRepositoryUser, altDeploymentRepositoryPassword, utils)
		if err != nil {
			return "", errors.Wrap(err, "Could not update settings xml")
		}
		return projectSettingsFilePath, nil
	} else {
		projectSettingsFilePath, err := maven.CreateNewProjectSettingsXML(altDeploymentRepositoryID, altDeploymentRepositoryUser, altDeploymentRepositoryPassword, utils)
		if err != nil {
			return "", errors.Wrap(err, "Could not create settings xml")
		}
		return projectSettingsFilePath, nil
	}
}

func loadRemoteRepoCertificates(certificateList []string, client piperhttp.Downloader, flags *[]string, runner command.ExecRunner, fileUtils piperutils.FileUtils, javaCaCertFilePath string) error {
	existingJavaCaCerts := filepath.Join(os.Getenv("JAVA_HOME"), "jre", "lib", "security", "cacerts")

	if len(javaCaCertFilePath) > 0 {
		existingJavaCaCerts = javaCaCertFilePath
	}

	exists, err := fileUtils.FileExists(existingJavaCaCerts)

	if err != nil {
		return errors.Wrap(err, "Could not find the existing java cacerts")
	}

	if !exists {
		return errors.Wrap(err, "Could not find the existing java cacerts")
	}

	trustStore := filepath.Join(getWorkingDirForTrustStore(), ".pipeline", "mavenCaCerts")

	log.Entry().Infof("copying java cacerts : %s to new cacerts : %s", existingJavaCaCerts, trustStore)
	_, fileUtilserr := fileUtils.Copy(existingJavaCaCerts, trustStore)

	if fileUtilserr != nil {
		return errors.Wrap(err, "Could not copy existing cacerts into new cacerts location ")
	}

	log.Entry().Infof("using trust store %s", trustStore)

	if exists, _ := fileUtils.FileExists(trustStore); exists {
		maven_opts := "-Djavax.net.ssl.trustStore=.pipeline/mavenCaCerts -Djavax.net.ssl.trustStorePassword=changeit"
		err := os.Setenv("MAVEN_OPTS", maven_opts)
		if err != nil {
			return errors.Wrap(err, "Could not create MAVEN_OPTS environment variable ")
		}
		log.Entry().WithField("trust store", trustStore).Info("Using local trust store")
	}

	if len(certificateList) > 0 {
		keytoolOptions := []string{
			"-import",
			"-noprompt",
			"-storepass", "changeit",
			"-keystore", trustStore,
		}
		tmpFolder := getTempDirForCertFile()
		defer os.RemoveAll(tmpFolder) // clean up

		for _, certificate := range certificateList {
			filename := path.Base(certificate) // decode?
			target := filepath.Join(tmpFolder, filename)

			log.Entry().WithField("source", certificate).WithField("target", target).Info("Downloading TLS certificate")
			// download certificate
			if err := client.DownloadFile(certificate, target, nil, nil); err != nil {
				return errors.Wrapf(err, "Download of TLS certificate failed")
			}
			options := append(keytoolOptions, "-file", target)
			options = append(options, "-alias", filename)
			// add certificate to keystore
			if err := runner.RunExecutable("keytool", options...); err != nil {
				return errors.Wrap(err, "Adding certificate to keystore failed")
			}
		}
		log.Entry().Infof("custom tls certificates successfully added to the trust store %s", trustStore)
	} else {
		log.Entry().Debug("Download of TLS certificates skipped")
	}
	return nil
}

func getWorkingDirForTrustStore() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Entry().WithError(err).WithField("path", workingDir).Debug("Retrieving of work directory failed")
	}
	return workingDir
}

func getTempDirForCertFile() string {
	tmpFolder, err := ioutil.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
	}
	return tmpFolder
}
