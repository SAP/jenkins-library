package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

func mavenBuild(config mavenBuildOptions, telemetryData *telemetry.CustomData) {
	utils := maven.NewUtilsBundle()

	err := runMavenBuild(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMavenBuild(config *mavenBuildOptions, telemetryData *telemetry.CustomData, utils maven.Utils) error {
	downloadClient := &piperhttp.Client{}

	var flags = []string{"-update-snapshots", "--batch-mode"}

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

			runner := &command.Command{}
			fileUtils := &piperutils.Files{}

			log.Entry().Infof("Successfully found deployment user : '%s'", config.AltDeploymentRepositoryUser)
			log.Entry().Infof("Successfully found deployment password : '%s'", config.AltDeploymentRepositoryPassowrd)
			log.Entry().Infof("Successfully found deployment id : '%s'", config.AltDeploymentRepositoryID)
			log.Entry().Infof("Successfully found deployment url : '%s'", config.AltDeploymentRepositoryURL)

			if (len(config.AltDeploymentRepositoryID) > 0) && (len(config.AltDeploymentRepositoryPassowrd) > 0) && (len(config.AltDeploymentRepositoryUser) > 0) {
				// update or create a new project settings xml
				if len(config.ProjectSettingsFile) > 0 {
					err := maven.UpdateProjectSettingsXML(config.ProjectSettingsFile, config.AltDeploymentRepositoryID, config.AltDeploymentRepositoryUser, config.AltDeploymentRepositoryPassowrd, utils)
					if err != nil {
						return err
					}
				} else {
					projectSettingsFile, err := maven.CreateNewProjectSettingsXML(config.AltDeploymentRepositoryID, config.AltDeploymentRepositoryUser, config.AltDeploymentRepositoryPassowrd, utils)
					if err != nil {
						return err
					} else {
						mavenOptions.ProjectSettingsFile = projectSettingsFile
					}
				}
			}

			deployFlags := []string{"-Dmaven.main.skip=true", "-Dmaven.test.skip=true", "-Dmaven.install.skip=true"}
			if (len(config.AltDeploymentRepositoryID) > 0) && (len(config.AltDeploymentRepositoryURL) > 0) {
				deployFlags = append(deployFlags, "-DaltDeploymentRepository="+config.AltDeploymentRepositoryID+"::default::"+config.AltDeploymentRepositoryURL)
			}

			if err := loadRemoteRepoCertificates(config.CustomTLSCertificateLinks, downloadClient, &deployFlags, runner, fileUtils); err != nil {
				log.SetErrorCategory(log.ErrorInfrastructure)
				return err
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
	return err
}

func loadRemoteRepoCertificates(certificateList []string, client piperhttp.Downloader, flags *[]string, runner command.ExecRunner, fileUtils piperutils.FileUtils) error {
	trustStore := filepath.Join(getWorkingDirForTrustStore(), ".pipeline", "keystore.jks")
	log.Entry().Infof("using trust store %s", trustStore)

	if exists, _ := fileUtils.FileExists(trustStore); exists {
		maven_opts := "-Djavax.net.ssl.trustStore=" + trustStore + " -Djavax.net.ssl.trustStorePassword=changeit"
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

		maven_opts := "-Djavax.net.ssl.trustStore=.pipeline/keystore.jks -Djavax.net.ssl.trustStorePassword=changeit"
		err := os.Setenv("MAVEN_OPTS", maven_opts)
		if err != nil {
			return errors.Wrap(err, "Could not create MAVEN_OPTS environment variable ")
		}
		log.Entry().WithField("trust store", trustStore).Info("Using local trust store")
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
