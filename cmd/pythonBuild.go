package cmd

import (
	"fmt"

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

func pythonBuild(config pythonBuildOptions, telemetryData *telemetry.CustomData) {
	utils := newPythonBuildUtils()

	err := runPythonBuild(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runPythonBuild(config *pythonBuildOptions, telemetryData *telemetry.CustomData, utils pythonBuildUtils) error {

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

	if config.Publish {
		if err := publishWithTwine(config, utils, installFlags); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	return nil
}

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils, installFlags []string) error {
	var flags []string
	flags = append(flags, config.BuildFlags...)
	flags = append(flags, "setup.py", "sdist", "bdist_wheel")

	log.Entry().Info("starting building python project:")
	err := utils.RunExecutable("python3", flags...)
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
