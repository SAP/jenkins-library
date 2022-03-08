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
		if err := utils.RunExecutable("python3", "-m", "pip", "install", "cyclonedx-bom"); err != nil {
			return fmt.Errorf("failed to install 'cyclonedx-bom': %w", err)
		}
		if err := runBOMCreationForPy(utils); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	if config.Publish {
		if err := utils.RunExecutable("python3", "-m", "pip", "install", "twine"); err != nil {
			return fmt.Errorf("failed to install 'twine': %w", err)
		}
		if err := publishWithTwine(utils); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	return nil
}

//python3 -m pip install cyclonedx-bom

func publishWithTwine(utils pythonBuildUtils) error {
	if err := utils.RunExecutable("twine", "upload", "dist/*"); err != nil {
		return err
	}
	return nil
}

func runBOMCreationForPy(utils pythonBuildUtils) error {
	if err := utils.RunExecutable("cyclonedx-bom", "mod", "-licenses", "-test", "-output", PyBomFilename); err != nil {
		return err
	}
	return nil
}

func buildExecute(config *pythonBuildOptions, utils pythonBuildUtils) error {
	var flags []string
	flags = append(flags, "-m", "build")

	if err := utils.RunExecutable("python3", "-m", "pip", "install", "--upgrade", "build"); err != nil {
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
