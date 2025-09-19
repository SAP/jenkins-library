package cmd

import (
	"fmt"

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

	if err := python.BuildWithSetupPy(utils.RunExecutable, config.VirtualEnvironmentName, config.BuildFlags, config.SetupFlags); err != nil {
		return err
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

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils, pipInstallFlags []string, virutalEnvironmentPathMap map[string]string) error {
	log.Entry().Info("starting building python project using pypa/build:")
	// Use the venv python to run the build backend; this creates sdist and wheel in dist/
	if err := utils.RunExecutable(virutalEnvironmentPathMap["python"], "-m", "build", "--sdist", "--wheel", "."); err != nil {
		return err
	}
	return nil
}

func createVirtualEnvironment(utils pythonBuildUtils, config *pythonBuildOptions, virtualEnvironmentPathMap map[string]string) error {
	virtualEnvironmentFlags := []string{"-m", "venv", config.VirutalEnvironmentName}
	if err := utils.RunExecutable("python3.12", virtualEnvironmentFlags...); err != nil {
		return err
	}
	// activating in a separate shell is not required for subsequent calls since we use explicit venv paths,
	// but keep the call to mirror previous behavior
	if err := utils.RunExecutable("bash", "-c", "source "+filepath.Join(config.VirutalEnvironmentName, "bin", "activate")); err != nil {
		return err
	}

	pipPath := filepath.Join(config.VirutalEnvironmentName, "bin", "pip")
	virtualEnvironmentPathMap["pip"] = pipPath
	// Use the venv's python binary so packages installed into the venv (like setuptools) are available
	virtualEnvironmentPathMap["python"] = filepath.Join(config.VirutalEnvironmentName, "bin", "python")
	virtualEnvironmentPathMap["deactivate"] = filepath.Join(config.VirutalEnvironmentName, "bin", "deactivate")

	// Upgrade pip and install build/wheel/setuptools into the virtual environment
	if err := utils.RunExecutable(pipPath, "install", "--upgrade", "pip", "build", "wheel", "setuptools"); err != nil {
		return err
	}

	return nil
}

func removeVirtualEnvironment(utils pythonBuildUtils, config *pythonBuildOptions) error {
	err := utils.RemoveAll(config.VirutalEnvironmentName)
	if err != nil {
		return err
	}
	return nil
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
