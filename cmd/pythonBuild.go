package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	PyBomFilename           = "bom-pip.xml"
	stepName                = "pythonBuild"
	cycloneDxPackageVersion = "cyclonedx-bom==3.11.0"
	cycloneDxSchemaVersion  = "1.4"
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
			StepName: "pythonBuild",
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

	pipInstallFlags := []string{"install", "--upgrade"}
	virtualEnvironmentPathMap := make(map[string]string)

	err := createVirtualEnvironment(utils, config, virtualEnvironmentPathMap)
	if err != nil {
		return err
	}

	err = buildExecute(config, utils, pipInstallFlags, virtualEnvironmentPathMap)
	if err != nil {
		return fmt.Errorf("Python build failed with error: %w", err)
	}

	if config.CreateBOM {
		if err := runBOMCreationForPy(utils, pipInstallFlags, virtualEnvironmentPathMap, config); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	log.Entry().Debugf("creating build settings information...")

	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return err
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
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

	if config.Publish {
		if err := publishWithTwine(config, utils, pipInstallFlags, virtualEnvironmentPathMap); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	err = removeVirtualEnvironment(utils, config)
	if err != nil {
		return err
	}

	return nil
}

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils, pipInstallFlags []string, virtualEnvironmentPathMap map[string]string) error {

	var flags []string
	flags = append(flags, config.BuildFlags...)
	flags = append(flags, "setup.py", "sdist", "bdist_wheel")

	log.Entry().Info("starting building python project:")
	err := utils.RunExecutable(virtualEnvironmentPathMap["python"], flags...)
	if err != nil {
		return err
	}
	return nil
}

func createVirtualEnvironment(utils pythonBuildUtils, config *pythonBuildOptions, virtualEnvironmentPathMap map[string]string) error {
	virtualEnvironmentFlags := []string{"-m", "venv", config.VirtualEnvironmentName}
	err := utils.RunExecutable("python3", virtualEnvironmentFlags...)
	if err != nil {
		return err
	}
	err = utils.RunExecutable("bash", "-c", "source "+filepath.Join(config.VirtualEnvironmentName, "bin", "activate"))
	if err != nil {
		return err
	}
	virtualEnvironmentPathMap["pip"] = filepath.Join(config.VirtualEnvironmentName, "bin", "pip")
	// venv will create symlinks to python3 inside the container
	virtualEnvironmentPathMap["python"] = "python"
	virtualEnvironmentPathMap["deactivate"] = filepath.Join(config.VirtualEnvironmentName, "bin", "deactivate")

	return nil
}

func removeVirtualEnvironment(utils pythonBuildUtils, config *pythonBuildOptions) error {
	err := utils.RemoveAll(config.VirtualEnvironmentName)
	if err != nil {
		return err
	}
	return nil
}

func runBOMCreationForPy(utils pythonBuildUtils, pipInstallFlags []string, virtualEnvironmentPathMap map[string]string, config *pythonBuildOptions) error {
	pipInstallFlags = append(pipInstallFlags, cycloneDxPackageVersion)
	if err := utils.RunExecutable(virtualEnvironmentPathMap["pip"], pipInstallFlags...); err != nil {
		return err
	}
	virtualEnvironmentPathMap["cyclonedx"] = filepath.Join(config.VirtualEnvironmentName, "bin", "cyclonedx-py")

	if err := utils.RunExecutable(virtualEnvironmentPathMap["cyclonedx"], "--e", "--output", PyBomFilename, "--format", "xml", "--schema-version", cycloneDxSchemaVersion); err != nil {
		return err
	}
	return nil
}

func publishWithTwine(config *pythonBuildOptions, utils pythonBuildUtils, pipInstallFlags []string, virtualEnvironmentPathMap map[string]string) error {
	pipInstallFlags = append(pipInstallFlags, "twine")
	if err := utils.RunExecutable(virtualEnvironmentPathMap["pip"], pipInstallFlags...); err != nil {
		return err
	}
	virtualEnvironmentPathMap["twine"] = filepath.Join(config.VirtualEnvironmentName, "bin", "twine")
	if err := utils.RunExecutable(virtualEnvironmentPathMap["twine"], "upload", "--username", config.TargetRepositoryUser,
		"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL, "--disable-progress-bar",
		"dist/*"); err != nil {
		return err
	}
	return nil
}
