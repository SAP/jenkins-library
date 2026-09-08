package cmd

import (
	"encoding/json"
	"fmt"
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
)

const (
	cycloneDxVersion       = "7.3.0"
	CycloneDxSchemaVersion = "1.4"
	stepName               = "pythonBuild"
)

type pythonBuildUtils interface {
	command.ExecRunner
	piperutils.FileUtils
}

type pythonBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
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
	exitHandler, err := python.CreateVirtualEnvironment(utils.RunExecutable, utils.RemoveAll, config.VirtualEnvironmentName)
	if err != nil {
		return err
	}
	log.DeferExitHandler(exitHandler)
	defer exitHandler()

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

	if config.RunTests {
		if err := python.InstallTestDependencies(utils.RunExecutable, config.VirtualEnvironmentName); err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return fmt.Errorf("failed to install test dependencies: %w", err)
		}
		if err := python.RunTests(utils.RunExecutable, config.VirtualEnvironmentName, config.TestOptions); err != nil {
			log.SetErrorCategory(log.ErrorTest)
			return fmt.Errorf("failed to run python tests: %w", err)
		}
	}

	if config.CreateBOM {
		if err := python.CreateBOM(utils.RunExecutable, utils.FileExists, utils.ReadFile, config.VirtualEnvironmentName, config.RequirementsFilePath, cycloneDxVersion, CycloneDxSchemaVersion); err != nil {
			return fmt.Errorf("failed to create BOM: %w", err)
		}
	}

	commonPipelineEnvironment.custom.buildSettingsInfo = createBuildSettingsInfo(config)

	if config.CreateBuildArtifactsMetadata {
		if err := createPythonBuildArtifactsMetadata(commonPipelineEnvironment); err != nil {
			log.Entry().Warnf("unable to create build artifact metadata: %v", err)
		}
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
	return nil
}

// TODO: extract to common place
// GetDockerImageValue failure is intentionally non-fatal: the docker image is only metadata
// used for build-settings tracking. A missing or unresolvable image does not affect the build
// output, so we warn and continue rather than aborting an otherwise successful build.
func createBuildSettingsInfo(config *pythonBuildOptions) string {
	log.Entry().Debugf("creating build settings information...")
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		log.Entry().Warnf("failed to retrieve docker image value: %v", err)
		dockerImage = ""
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
	return buildSettingsInfo
}

func createPythonBuildArtifactsMetadata(commonPipelineEnvironment *pythonBuildCommonPipelineEnvironment) error {
	options := versioning.Options{}
	artifact, err := versioning.GetArtifact("pip", "", &options, nil)
	if err != nil {
		return fmt.Errorf("failed to get artifact: %w", err)
	}
	coordinate, err := artifact.GetCoordinates()
	if err != nil {
		return fmt.Errorf("failed to get artifact coordinates: %w", err)
	}

	component := piperutils.GetComponent(python.BOMFilename)
	coordinate.PURL = component.Purl

	var buildArtifacts build.BuildArtifacts
	buildArtifacts.Coordinates = []versioning.Coordinates{coordinate}
	jsonResult, err := json.Marshal(buildArtifacts)
	if err != nil {
		return fmt.Errorf("failed to marshal build artifacts: %w", err)
	}
	commonPipelineEnvironment.custom.pythonBuildArtifacts = string(jsonResult)
	return nil
}

func searchDescriptor(supported []string, existsFunc func(string) (bool, error)) (string, error) {
	var descriptor string
	for _, f := range supported {
		exists, err := existsFunc(f)
		if err != nil {
			return "", fmt.Errorf("checking %s: %w", f, err)
		}
		if exists {
			descriptor = f
			break
		}
	}
	if descriptor == "" {
		return "", fmt.Errorf("no build descriptor available, supported: %v", supported)
	}
	return descriptor, nil
}
