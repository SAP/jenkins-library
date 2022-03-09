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

var installFlags = []string{"-m", "pip", "install", "--upgrade"}

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

	tomlExists, err := utils.FileExists("pyproject.toml")
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		fmt.Errorf("failed to check for important file: %w", err)
	}
	if !tomlExists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("cannot run without important file")
	}

	err = buildExecute(config, utils)

	if config.CreateBOM {
		if err := runBOMCreationForPy(utils); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	if config.Publish {
		if err := publishWithTwine(config, utils); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	return nil
}

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils) error {
	var flags []string
	flags = append(flags, "-m", "build")
	installFlags = append(installFlags, "build")

	if err := utils.RunExecutable("python3", installFlags...); err != nil {
		return fmt.Errorf("failed to install 'build': %w", err)
	}
	flags = append(flags, config.BuildFlags...)

	setupPyExists, _ := utils.FileExists("setup.py")
	setupCFGExists, _ := utils.FileExists("setup.cfg")
	if setupPyExists || setupCFGExists {
		log.Entry().Info("starting building python project:")
		err := utils.RunExecutable("python3", flags...)
		if err != nil {
			log.Entry().Errorln("starting building python project can't start:", err)
		}
	}

	return nil
}

func runBOMCreationForPy(utils pythonBuildUtils) error {
	installFlags[len(installFlags)-1] = "cyclonedx-bom"
	if err := utils.RunExecutable("python3", installFlags...); err != nil {
		return err
	}
	if err := utils.RunExecutable("cyclonedx-bom", "--e", "--output", PyBomFilename); err != nil {
		return err
	}
	return nil
}

func publishWithTwine(config *pythonBuildOptions, utils pythonBuildUtils) error {
	installFlags[len(installFlags)-1] = "twine"
	if err := utils.RunExecutable("python3", installFlags...); err != nil {
		return err
	}
	if err := utils.RunExecutable("twine", "upload", "--username", config.TargetRepositoryUser,
		"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL,
		"dist/*.tar.gz"); err != nil {
		return err
	}
	return nil
}
