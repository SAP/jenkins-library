package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/feature"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/python"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	PyBomFilename           = "bom-pip.xml"
	stepName                = "pythonBuild"
	cycloneDxVersion        = "6.1.1"
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
	virutalEnvironmentPathMap := make(map[string]string)

	// create virtualEnv
	if err := createVirtualEnvironment(utils, config, virutalEnvironmentPathMap); err != nil {
		return err
	}
	//TODO: use a defer func to cleanup the virtual environment

	// FEATURE FLAG (com_sap_piper_featureFlag_pythonToml) to switch to new implementation of python build step
	if config.UseTomlFile || feature.IsFeatureEnabled("pythonToml") {
		// check project descriptor
		buildDescriptorFilePath, err := searchDescriptor([]string{"pyproject.toml", "setup.py"}, utils.FileExists)
		if err != nil {
			return err
		}
		// build package
		if strings.HasSuffix(buildDescriptorFilePath, "pyproject.toml") {
			if err := python.InstallProjectDependencies(utils.RunExecutable, python.Binary); err != nil {
				return fmt.Errorf("Failed to install project dependencies: %w", err)
			}
			if err := python.Build(utils.RunExecutable, python.Binary, config.BuildFlags, config.SetupFlags); err != nil {
				return fmt.Errorf("Failed to build python project: %w", err)
			}
		}
	} else {
		if err := buildExecute(config, utils, virutalEnvironmentPathMap); err != nil {
			return fmt.Errorf("Python build failed with error: %w", err)
		}
	}

	// generate BOM
	if config.CreateBOM {
		if err := runBOMCreationForPy(utils, virutalEnvironmentPathMap, config); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	// generate build settings information
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

	// publish package
	if config.Publish {
		if err := publishWithTwine(config, utils, virutalEnvironmentPathMap); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	// remove virtualEnv
	return removeVirtualEnvironment(utils, config)
}

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils, virutalEnvironmentPathMap map[string]string) error {
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
	if err := utils.RemoveAll(config.VirutalEnvironmentName); err != nil {
		return fmt.Errorf("failed to remove virtual environment: %w", err)
	}
	return nil
}

func runBOMCreationForPy(utils pythonBuildUtils, virutalEnvironmentPathMap map[string]string, config *pythonBuildOptions) error {
	// install dependencies from requirements.txt
	exists, _ := utils.FileExists(config.RequirementsFilePath)
	if exists {
		pipInstallRequirementsFlags := append(python.PipInstallFlags, "--requirement", config.RequirementsFilePath)
		if err := utils.RunExecutable(virutalEnvironmentPathMap["pip"], pipInstallRequirementsFlags...); err != nil {
			return err
		}
	} else {
		log.Entry().Warnf("unable to find requirements.txt file at %s , continuing SBOM generation without requirements.txt", config.RequirementsFilePath)
	}

	// install cyclonedx
	pipInstallCycloneDxFlags := append(python.PipInstallFlags, cycloneDxPackageVersion)
	if err := utils.RunExecutable(virutalEnvironmentPathMap["pip"], pipInstallCycloneDxFlags...); err != nil {
		return err
	}
	virutalEnvironmentPathMap["cyclonedx"] = filepath.Join(config.VirutalEnvironmentName, "bin", "cyclonedx-py")

	// run cyclonedx
	// TODO: use modules, python -m cyclonedx_py ... to avoid virutalEnvironmentPathMap
	if err := utils.RunExecutable(
		virutalEnvironmentPathMap["cyclonedx"],
		"env",
		"--output-file", PyBomFilename,
		"--output-format", "XML",
		"--spec-version", cycloneDxSchemaVersion,
	); err != nil {
		return err
	}
	return nil
}

func publishWithTwine(config *pythonBuildOptions, utils pythonBuildUtils, virutalEnvironmentPathMap map[string]string) error {
	if err := python.Install(utils.RunExecutable, "twine", "", config.VirutalEnvironmentName, virutalEnvironmentPathMap); err != nil {
		return err
	}

	// TODO: use modules, python -m twine ... to avoid virutalEnvironmentPathMap
	if err := utils.RunExecutable(
		virutalEnvironmentPathMap["twine"],
		"upload",
		"--username", config.TargetRepositoryUser,
		"--password", config.TargetRepositoryPassword,
		"--repository-url", config.TargetRepositoryURL,
		"--disable-progress-bar",
		"dist/*",
	); err != nil {
		return err
	}
	return nil
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
