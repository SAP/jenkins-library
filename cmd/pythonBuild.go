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
	PyBomFilename = "bom.xml"
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
		Command: &command.Command{},
		Files:   &piperutils.Files{},
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

	installFlags := []string{"-m", "pip", "install", "--upgrade"}

	err := buildExecute(config, utils, installFlags)
	if err != nil {
		return fmt.Errorf("Python build failed with error: %w", err)
	}

	if config.CreateBOM {
		if err := runBOMCreationForPy(utils, installFlags); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	log.Entry().Debugf("creating build settings information...")
	stepName := "pythonBuild"
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
		if err := publishWithTwine(config, utils, installFlags); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	return nil
}

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils, installFlags []string) error {
	err := createVirtualEnvironment(utils)
	if err != nil {
		return err
	}
	var flags []string
	flags = append(flags, config.BuildFlags...)
	flags = append(flags, "setup.py", "sdist", "bdist_wheel")

	log.Entry().Info("starting building python project:")
	err = utils.RunExecutable("python3", flags...)
	if err != nil {
		return err
	}

	return nil
}

func createVirtualEnvironment(utils pythonBuildUtils) error {
	virtualEnvironmentFlags := []string{"-m", "venv", "piperBuild-env"}
	err := utils.RunExecutable("python3", virtualEnvironmentFlags...)
	if err != nil {
		return err
	}
	err = utils.RunExecutable("bash", "-c", "source", filepath.Join("piperBuild-env", "bin", "activate"))
	if err != nil {
		return err
	}
	return nil
}

func runBOMCreationForPy(utils pythonBuildUtils, installFlags []string) error {
	installFlags = append(installFlags, "cyclonedx-bom")
	if err := utils.RunExecutable("python3", installFlags...); err != nil {
		return err
	}
	if err := utils.RunExecutable("cyclonedx-bom", "--e", "--output", PyBomFilename); err != nil {
		return err
	}
	return nil
}

func publishWithTwine(config *pythonBuildOptions, utils pythonBuildUtils, installFlags []string) error {
	installFlags = append(installFlags, "twine")
	if err := utils.RunExecutable("python3", installFlags...); err != nil {
		return err
	}
	if err := utils.RunExecutable("twine", "upload", "--username", config.TargetRepositoryUser,
		"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL,
		"dist/*"); err != nil {
		return err
	}
	return nil
}
