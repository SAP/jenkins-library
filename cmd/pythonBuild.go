package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/python"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
	if exitHandler, err := python.CreateVirtualEnvironment(utils.RunExecutable, utils.RemoveAll, config.VirtualEnvironmentName); err != nil {
		return err
	} else {
		log.DeferExitHandler(exitHandler)
		defer exitHandler()
	}

	// check project descriptor
	buildDescriptorFilePath, err := searchDescriptor([]string{"pyproject.toml", "setup.py"}, utils.FileExists)
	if err != nil {
		return err
	}

	if strings.HasSuffix(buildDescriptorFilePath, "pyproject.toml") {
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}
		utils.AppendEnv([]string{
			fmt.Sprintf("VIRTUAL_ENV=%s", filepath.Join(workDir, config.VirtualEnvironmentName)),
		})
		// handle pyproject.toml file
		if err := python.InstallPip(virtualEnvPathMap["pip"], utils.RunExecutable); err != nil {
			return fmt.Errorf("failed to upgrade pip: %w", err)
		}
		if err := python.InstallProjectDependencies(virtualEnvPathMap["pip"], utils.RunExecutable); err != nil {
			return fmt.Errorf("failed to install project dependencies: %w", err)
		}
		if err := python.InstallBuild(virtualEnvPathMap["pip"], utils.RunExecutable); err != nil {
			return fmt.Errorf("failed to install build module: %w", err)
		}
		if err := python.InstallWheel(virtualEnvPathMap["pip"], utils.RunExecutable); err != nil {
			return fmt.Errorf("failed to install wheel module: %w", err)
		}
		if err := python.Build(virtualEnvPathMap["python"], utils.RunExecutable, config.BuildFlags, config.SetupFlags); err != nil {
			return fmt.Errorf("failed to build python project: %w", err)
		}
	} else {
		// handle legacy setup.py file
		if err := python.BuildWithSetupPy(utils.RunExecutable, config.VirtualEnvironmentName, config.BuildFlags, config.SetupFlags); err != nil {
			return fmt.Errorf("failed to build python project: %w", err)
		}
	}

	if config.CreateBOM {
		if err := python.CreateBOM(utils.RunExecutable, utils.FileExists, config.VirtualEnvironmentName, config.RequirementsFilePath, cycloneDxVersion, cycloneDxSchemaVersion); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	if info, err := createBuildSettingsInfo(config); err != nil {
		return err
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
	return nil
}

// TODO: extract to common place
func createBuildSettingsInfo(config *pythonBuildOptions) (string, error) {
	// generate build settings information
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
