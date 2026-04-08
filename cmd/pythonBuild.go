package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/build"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/python"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

const (
	cycloneDxVersion       = "6.1.1"
	cycloneDxSchemaVersion = "1.4"
	stepName               = "pythonBuild"
)

type pythonBuildUtils interface {
	command.ExecRunner
	FileExists(filename string) (bool, error)
	piperutils.FileUtils

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

type pythonBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	httpClient *piperhttp.Client
}

func newPythonBuildUtils() pythonBuildUtils {
	utils := pythonBuildUtilsBundle{
		Command: &command.Command{
			StepName: stepName,
		},
		Files: &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func pythonBuild(config pythonBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *pythonBuildCommonPipelineEnvironment) {
	utils := newPythonBuildUtils()

	err := runPythonBuild(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runPythonBuild(config *pythonBuildOptions, telemetryData *telemetry.CustomData, utils pythonBuildUtils, commonPipelineEnvironment *pythonBuildCommonPipelineEnvironment) error {
	if exitHandler, err := python.CreateVirtualEnvironment(utils.RunExecutable, utils.RemoveAll, config.VirtualEnvironmentName); err != nil {
		return err
	} else {
		log.DeferExitHandler(exitHandler)
		defer exitHandler()
	}

	// check project descriptor
	buildDescriptorFilePath, err := searchDescriptor([]string{"pyproject.toml", "setup.py"}, utils.FileExists)
	if err != nil {
		return fmt.Errorf("failed to determine build descriptor file: %w", err)
	}

	if strings.HasSuffix(buildDescriptorFilePath, "pyproject.toml") {
		// handle pyproject.toml file
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}
		utils.AppendEnv([]string{
			fmt.Sprintf("VIRTUAL_ENV=%s", filepath.Join(workDir, config.VirtualEnvironmentName)),
		})
		if err := python.BuildWithPyProjectToml(utils.RunExecutable, config.VirtualEnvironmentName, config.BuildFlags, config.SetupFlags); err != nil {
			return fmt.Errorf("failed to build python project: %w", err)
		}
	} else {
		// handle legacy setup.py file
		if err := python.BuildWithSetupPy(utils.RunExecutable, config.VirtualEnvironmentName, config.BuildFlags, config.SetupFlags); err != nil {
			return fmt.Errorf("failed to build python project: %w", err)
		}
	}

	// coordinate contains the artifact id and version needed for sbom generation when only
	// setup.py is present as project descriptor
	err, coordinate := createPythonBuildArtifactsMetadata(buildDescriptorFilePath, config.TargetRepositoryURL, utils)
	if err != nil {
		log.Entry().Warnf("unable to create build artifact metadata : %v", err)
	}

	if config.CreateBOM {
		if err := python.CreateBOM(utils.RunExecutable, utils.FileExists, utils.ReadFile, config.VirtualEnvironmentName, config.RequirementsFilePath, cycloneDxVersion, cycloneDxSchemaVersion, buildDescriptorFilePath, coordinate); err != nil {
			return fmt.Errorf("failed to create BOM: %w", err)
		}
	}

	component := piperutils.GetComponent(filepath.Join(filepath.Dir(buildDescriptorFilePath), python.BOMFilename))

	purl := component.Purl
	coordinate.PURL = purl

	if info, err := createBuildSettingsInfo(config); err != nil {
		return fmt.Errorf("failed to create build settings info: %v", err)
	} else {
		commonPipelineEnvironment.custom.buildSettingsInfo = info
	}

	if config.Publish {
		if err := python.PublishPackage(
			utils.RunExecutable,
			config.VirtualEnvironmentName,
			config.TargetRepositoryURL,
			config.TargetRepositoryUser,
			config.TargetRepositoryPassword,
		); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	if config.CreateBuildArtifactsMetadata {
		var buildArtifacts build.BuildArtifacts
		buildArtifacts.Coordinates = append(buildArtifacts.Coordinates, coordinate)
		jsonResult, _ := json.Marshal(buildArtifacts)
		commonPipelineEnvironment.custom.pythonBuildArtifacts = string(jsonResult)
	} else {
		log.Entry().Info("skipping creation of build artifacts metadata")
	}

	return nil
}

func createPythonBuildArtifactsMetadata(buildDescriptorFilePath string, targetRepositoryURL string, utils pythonBuildUtils) (error, versioning.Coordinates) {
	options := versioning.Options{}
	builtArtifact, err := versioning.GetArtifact("pip", buildDescriptorFilePath, &options, utils)
	if err != nil {
		return err, versioning.Coordinates{}
	}
	coordinate, err := builtArtifact.GetCoordinates()
	if err != nil {
		return err, versioning.Coordinates{}
	}

	coordinate.URL = targetRepositoryURL
	coordinate.BuildPath = filepath.Dir(buildDescriptorFilePath)

	return nil, coordinate
}

// TODO: extract to common place
func createBuildSettingsInfo(config *pythonBuildOptions) (string, error) {
	log.Entry().Debugf("creating build settings information...")
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return "", err
	}
	pythonConfig := buildsettings.BuildOptions{
		CreateBOM:         config.CreateBOM,
		Publish:           config.Publish,
		BuildSettingsInfo: config.BuildSettingsInfo,
		DockerImage:       dockerImage,
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&pythonConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	return buildSettingsInfo, nil
}

func searchDescriptor(supported []string, existsFunc func(string) (bool, error)) (string, error) {
	var descriptor string
	for _, f := range supported {
		exists, _ := existsFunc(f)
		if exists {
			descriptor = f
			break
		}
	}
	if len(descriptor) == 0 {
		return "", fmt.Errorf("no build descriptor available, supported: %v", supported)
	}
	return descriptor, nil
}

func (p *pythonBuildUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return p.httpClient.DownloadFile(url, filename, header, cookies)
}
