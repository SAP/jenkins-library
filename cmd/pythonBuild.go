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
	cycloneDxPackageVersion = "cyclonedx-bom==6.1.1"
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
	virutalEnvironmentPathMap := make(map[string]string)

	err := createVirtualEnvironment(utils, config, virutalEnvironmentPathMap)
	if err != nil {
		return err
	}

	err = buildExecute(config, utils, pipInstallFlags, virutalEnvironmentPathMap)
	if err != nil {
		return fmt.Errorf("Python build failed with error: %w", err)
	}

	if config.CreateBOM {
		if err := runBOMCreationForPy(utils, pipInstallFlags, virutalEnvironmentPathMap, config); err != nil {
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
		if err := publishWithTwine(config, utils, pipInstallFlags, virutalEnvironmentPathMap); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	err = removeVirtualEnvironment(utils, config)
	if err != nil {
		return err
	}

	return nil
}

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils, pipInstallFlags []string, virutalEnvironmentPathMap map[string]string) error {
	var flags []string
	flags = append(flags, config.BuildFlags...)
	flags = append(flags, "setup.py")
	flags = append(flags, config.SetupFlags...)
	flags = append(flags, "sdist", "bdist_wheel")

	log.Entry().Info("starting building python project:")
	err := utils.RunExecutable(virutalEnvironmentPathMap["python"], flags...)
	if err != nil {
		return err
	}
	return nil
}

func createVirtualEnvironment(utils pythonBuildUtils, config *pythonBuildOptions, virutalEnvironmentPathMap map[string]string) error {
	virtualEnvironmentFlags := []string{"-m", "venv", config.VirutalEnvironmentName}
	err := utils.RunExecutable("python3", virtualEnvironmentFlags...)
	if err != nil {
		return err
	}
	err = utils.RunExecutable("bash", "-c", "source "+filepath.Join(config.VirutalEnvironmentName, "bin", "activate"))
	if err != nil {
		return err
	}
	virutalEnvironmentPathMap["pip"] = filepath.Join(config.VirutalEnvironmentName, "bin", "pip")
	// venv will create symlinks to python3 inside the container
	virutalEnvironmentPathMap["python"] = "python"
	virutalEnvironmentPathMap["deactivate"] = filepath.Join(config.VirutalEnvironmentName, "bin", "deactivate")

	return nil
}

func removeVirtualEnvironment(utils pythonBuildUtils, config *pythonBuildOptions) error {
	err := utils.RemoveAll(config.VirutalEnvironmentName)
	if err != nil {
		return err
	}
	return nil
}

func runBOMCreationForPy(utils pythonBuildUtils, pipInstallFlags []string, virutalEnvironmentPathMap map[string]string, config *pythonBuildOptions) error {
	pipInstallOriginalFlags := pipInstallFlags
	exists, _ := utils.FileExists(config.RequirementsFilePath)
	if exists {
		pipInstallRequirementsFlags := append(pipInstallOriginalFlags, "--requirement", config.RequirementsFilePath)
		if err := utils.RunExecutable(virutalEnvironmentPathMap["pip"], pipInstallRequirementsFlags...); err != nil {
			return err
		}
	} else {
		log.Entry().Warnf("unable to find requirements.txt file at %s , continuing SBOM generation without requirements.txt", config.RequirementsFilePath)
	}

	pipInstallCycloneDxFlags := append(pipInstallOriginalFlags, cycloneDxPackageVersion)

	if err := utils.RunExecutable(virutalEnvironmentPathMap["pip"], pipInstallCycloneDxFlags...); err != nil {
		return err
	}
	virutalEnvironmentPathMap["cyclonedx"] = filepath.Join(config.VirutalEnvironmentName, "bin", "cyclonedx-py")

	if err := utils.RunExecutable(virutalEnvironmentPathMap["cyclonedx"], "--e", "--output", PyBomFilename, "--format", "xml", "--schema-version", cycloneDxSchemaVersion); err != nil {
		return err
	}
	return nil
}

func publishWithTwine(config *pythonBuildOptions, utils pythonBuildUtils, pipInstallFlags []string, virutalEnvironmentPathMap map[string]string) error {
	pipInstallFlags = append(pipInstallFlags, "twine")
	if err := utils.RunExecutable(virutalEnvironmentPathMap["pip"], pipInstallFlags...); err != nil {
		return err
	}
	virutalEnvironmentPathMap["twine"] = filepath.Join(config.VirutalEnvironmentName, "bin", "twine")
	if err := utils.RunExecutable(virutalEnvironmentPathMap["twine"], "upload", "--username", config.TargetRepositoryUser,
		"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL, "--disable-progress-bar",
		"dist/*"); err != nil {
		return err
	}
	return nil
}
