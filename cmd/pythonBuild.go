package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"os"
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

	if len(config.Sources) != 0 {
		for _, source := range config.Sources {
			exists, err := utils.DirExists(source)
			if err != nil {
				log.Entry().WithError(err).Error("failed to check for python project dir")
				return fmt.Errorf("failed to check for python project dir: %w", err)
			}
			if !exists {
				log.Entry().WithError(err).Errorf("the python project dir '%v' could not be found: %v", source, err)
				return fmt.Errorf("the python project dir '%v' could not be found", source)
			} else {
				tomlExists, err := utils.FileExists(source + "/pyproject.toml")
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("failed to check for important file: %w", err)
				}
				if !tomlExists {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("cannot run without important file")
				}
			}
			err = buildExecute(source, config, utils)
			if err != nil {

			}
		}
	} else {
		source, err := os.Getwd()
		if err != nil {
			log.Entry().WithError(err).Error("Getting current working directory failed.")
		}

		err = buildExecute(source, config, utils)
	}

	return nil
}

func buildExecute(source string, config *pythonBuildOptions, utils pythonBuildUtils) error {
	var flags []string
	flags = append(flags, "-m", "build")

	if err := utils.RunExecutable("python3", "-m", "pip", "install", "--upgrade", "build"); err != nil {
		return fmt.Errorf("failed to install 'build': %w", err)
	}
	flags = append(flags, source)
	flags = append(flags, config.BuildFlags...)

	log.Entry().Info("starting building python project:", source)
	err := utils.RunExecutable("python3", flags...)
	if err != nil {
		log.Entry().Errorln("starting building python project can't start:", err)
	}
	return nil
}
