package cmd

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/SAP/jenkins-library/pkg/build"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/pkg/errors"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

const (
	mvnBomFilename       = "bom-maven"
	mvnSimpleBomFilename = "simple-bom-maven"
	mvnCycloneDXPackage  = "org.cyclonedx:cyclonedx-maven-plugin:2.9.1"
)

func mavenBuild(config mavenBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *mavenBuildCommonPipelineEnvironment) {
	utils := maven.NewUtilsBundle()
	// enables url-log.json creation
	cmd := reflect.ValueOf(utils).Elem().FieldByName("Command")
	if cmd.IsValid() {
		reflect.Indirect(cmd).FieldByName("StepName").SetString("mavenBuild")
	}

	err := runMavenBuild(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMakeBOMGoal(config *mavenBuildOptions, utils maven.Utils) error {
	flags := []string{"--batch-mode"}
	// decision for BOM generation
	if shouldUpdateSnapshots(utils, config) {
		flags = append(flags, "-update-snapshots")
	} else {
		flags = append(flags, "--offline")
	}

	if len(config.Profiles) > 0 {
		flags = append(flags, "--activate-profiles", strings.Join(config.Profiles, ","))
	}
	exists, _ := utils.FileExists("integration-tests/pom.xml")
	if exists {
		flags = append(flags, "-pl", "!integration-tests")
	}

	var defines []string

	createBOMConfig := []string{
		"-DschemaVersion=1.4",
		"-DincludeBomSerialNumber=true",
		"-DincludeCompileScope=true",
		"-DincludeProvidedScope=true",
		"-DincludeRuntimeScope=true",
		"-DincludeSystemScope=true",
		"-DincludeTestScope=false",
		"-DincludeLicenseText=false",
		"-DoutputFormat=xml",
		"-DoutputName=" + mvnSimpleBomFilename,
	}
	defines = append(defines, createBOMConfig...)

	goals := []string{mvnCycloneDXPackage + ":makeBom"}

	if config.Flatten {
		goals = append(goals, "flatten:flatten")
		defines = append(defines, "-Dflatten.mode=resolveCiFriendliesOnly", "-DupdatePomFile=true")
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
	return err
}
func runMavenBuild(config *mavenBuildOptions, _ *telemetry.CustomData, utils maven.Utils, commonPipelineEnvironment *mavenBuildCommonPipelineEnvironment) error {
	flags := []string{"--batch-mode"}

	// Intelligent decision on snapshot updates
	useOfflineMode := !shouldUpdateSnapshots(utils, config)

	if useOfflineMode {
		flags = append(flags, "--offline")
		log.Entry().Info("Running in offline mode with cached dependencies")
	} else {
		flags = append(flags, "-update-snapshots")
	}

	if len(config.Profiles) > 0 {
		flags = append(flags, "--activate-profiles", strings.Join(config.Profiles, ","))
	}

	exists, _ := utils.FileExists("integration-tests/pom.xml")
	if exists {
		flags = append(flags, "-pl", "!integration-tests")
	}

	var defines []string
	var goals []string

	if config.Flatten {
		goals = append(goals, "flatten:flatten")
		defines = append(defines, "-Dflatten.mode=resolveCiFriendliesOnly", "-DupdatePomFile=true")
	}

	// DON'T add BOM goal to main build if we're in offline mode
	if config.CreateBOM && !useOfflineMode {
		// Only add BOM to main build if we're online
		goals = append(goals, mvnCycloneDXPackage+":makeAggregateBom")
		createBOMConfig := []string{
			"-DschemaVersion=1.4",
			"-DincludeBomSerialNumber=true",
			"-DincludeCompileScope=true",
			"-DincludeProvidedScope=true",
			"-DincludeRuntimeScope=true",
			"-DincludeSystemScope=true",
			"-DincludeTestScope=false",
			"-DincludeLicenseText=false",
			"-DoutputFormat=xml",
			"-DoutputName=" + mvnBomFilename,
		}
		defines = append(defines, createBOMConfig...)
	}

	goals = append(goals, "org.jacoco:jacoco-maven-plugin:prepare-agent")

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
	if err != nil {
		return errors.Wrapf(err, "failed to execute maven build for goal(s) '%v'", goals)
	}

	// If CreateBOM is true AND we ran in offline mode, run BOM generation separately in online mode
	if config.CreateBOM && useOfflineMode {
		log.Entry().Info("Running BOM generation separately in online mode")

		// Run just the BOM goals in online mode
		bomFlags := []string{"--batch-mode"} // No --offline flag

		if len(config.Profiles) > 0 {
			bomFlags = append(bomFlags, "--activate-profiles", strings.Join(config.Profiles, ","))
		}

		if exists {
			bomFlags = append(bomFlags, "-pl", "!integration-tests")
		}

		bomGoals := []string{mvnCycloneDXPackage + ":makeAggregateBom"}
		bomDefines := []string{
			"-DschemaVersion=1.4",
			"-DincludeBomSerialNumber=true",
			"-DincludeCompileScope=true",
			"-DincludeProvidedScope=true",
			"-DincludeRuntimeScope=true",
			"-DincludeSystemScope=true",
			"-DincludeTestScope=false",
			"-DincludeLicenseText=false",
			"-DoutputFormat=xml",
			"-DoutputName=" + mvnBomFilename,
		}

		bomMavenOptions := maven.ExecuteOptions{
			Flags:                       bomFlags,
			Goals:                       bomGoals,
			Defines:                     bomDefines,
			PomPath:                     config.PomPath,
			ProjectSettingsFile:         config.ProjectSettingsFile,
			GlobalSettingsFile:          config.GlobalSettingsFile,
			M2Path:                      config.M2Path,
			LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		}

		_, err := maven.Execute(&bomMavenOptions, utils)
		if err != nil {
			return errors.Wrap(err, "failed to execute BOM generation")
		}
	}

	if config.CreateBOM {
		// Separate run for makeBOM goal (simple BOM)
		if err := runMakeBOMGoal(config, utils); err != nil {
			return errors.Wrap(err, "failed to execute makeBOM goal")
		}
	}

	log.Entry().Debugf("creating build settings information...")
	stepName := "mavenBuild"
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return err
	}

	mavenConfig := buildsettings.BuildOptions{
		Profiles:                    config.Profiles,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		CreateBOM:                   config.CreateBOM,
		Publish:                     config.Publish,
		BuildSettingsInfo:           config.BuildSettingsInfo,
		DockerImage:                 dockerImage,
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&mavenConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

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

			deployFlags := []string{"--batch-mode"}

			if shouldUpdateSnapshots(utils, config) {
				deployFlags = append(deployFlags, "-update-snapshots")
			} else {
				deployFlags = append(deployFlags, "--offline")
			}

			if len(config.DeployFlags) > 0 {
				deployFlags = append(deployFlags, config.DeployFlags...)
			}
			if (len(config.AltDeploymentRepositoryID) > 0) && (len(config.AltDeploymentRepositoryURL) > 0) {
				deployFlags = append(deployFlags, "-DaltDeploymentRepository="+config.AltDeploymentRepositoryID+"::default::"+config.AltDeploymentRepositoryURL)
			}

			downloadClient := &piperhttp.Client{}
			downloadClient.SetOptions(piperhttp.ClientOptions{})
			runner := &command.Command{
				StepName: "mavenBuild",
			}
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
			if err != nil {
				return err
			}
			if config.CreateBuildArtifactsMetadata {
				createBuildArtifactsMetadata(config, commonPipelineEnvironment)
			}

			return nil
		} else {
			log.Entry().Infof("publish not detected, ignoring maven deploy")
		}
	}

	return err
}

func createBuildArtifactsMetadata(config *mavenBuildOptions, commonPipelineEnvironment *mavenBuildCommonPipelineEnvironment) bool {
	fileUtils := &piperutils.Files{}
	buildCoordinates := []versioning.Coordinates{}
	options := versioning.Options{
		ProjectSettingsFile: config.ProjectSettingsFile,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		M2Path:              config.M2Path,
	}
	var utils versioning.Utils

	matches, _ := fileUtils.Glob("**/pom.xml")
	for _, match := range matches {

		artifact, err := versioning.GetArtifact("maven", match, &options, utils)
		if err != nil {
			log.Entry().Warnf("unable to get artifact metdata : %v", err)
		} else {
			coordinate, err := artifact.GetCoordinates()
			if err != nil {
				log.Entry().Warnf("unable to get artifact coordinates : %v", err)
			} else {
				coordinate.BuildPath = filepath.Dir(match)
				coordinate.URL = config.AltDeploymentRepositoryURL
				coordinate.PURL = piperutils.GetPurl(filepath.Join(filepath.Dir(match), "/target/"+mvnSimpleBomFilename+".xml"))
				buildCoordinates = append(buildCoordinates, coordinate)
			}
		}
	}

	if len(buildCoordinates) == 0 {
		log.Entry().Warnf("unable to identify artifact coordinates for the maven packages published")
		return true
	}

	var buildArtifacts build.BuildArtifacts

	buildArtifacts.Coordinates = buildCoordinates
	jsonResult, _ := json.Marshal(buildArtifacts)
	commonPipelineEnvironment.custom.mavenBuildArtifacts = string(jsonResult)
	return false
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
	// TODO: make use of java/keytool package
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

	trustStore := filepath.Join(".pipeline", "mavenCaCerts")

	log.Entry().Infof("copying java cacerts : %s to new cacerts : %s", existingJavaCaCerts, trustStore)
	_, fileUtilserr := fileUtils.Copy(existingJavaCaCerts, trustStore)

	if fileUtilserr != nil {
		return errors.Wrap(err, "Could not copy existing cacerts into new cacerts location ")
	}

	if err := fileUtils.Chmod(trustStore, 0666); err != nil {
		return errors.Wrap(err, "unable to provide correct permission to trust store")
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

func getTempDirForCertFile() string {
	tmpFolder, err := os.MkdirTemp(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
	}
	return tmpFolder
}

func shouldUpdateSnapshots(utils maven.Utils, config *mavenBuildOptions) bool {
	cacheRestored := os.Getenv("PIPER_CACHE_RESTORED") == "true"
	dependenciesChanged := os.Getenv("PIPER_DEPENDENCIES_CHANGED") == "true"

	log.Entry().Debugf("Cache state - restored: %v, dependencies changed: %v", cacheRestored, dependenciesChanged)

	if !cacheRestored {
		log.Entry().Info("Cache not restored, will update snapshots")
		return true
	}

	if dependenciesChanged {
		log.Entry().Info("Dependencies changed, will update snapshots")
		return true
	}

	// Check if project actually uses SNAPSHOT dependencies
	pomContent, err := utils.FileRead(config.PomPath)
	if err != nil {
		log.Entry().Warnf("Could not read pom.xml to check for SNAPSHOTs: %v", err)
		return true // safe default
	}

	if strings.Contains(string(pomContent), "-SNAPSHOT</version>") {
		log.Entry().Info("SNAPSHOT dependencies detected, will check for updates")
		return true
	}

	log.Entry().Info("No SNAPSHOT dependencies found, skipping update checks")
	return false
}
