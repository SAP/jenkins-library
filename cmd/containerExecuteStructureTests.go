package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type containerExecuteStructureTestsUtils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
	Glob(pattern string) (matches []string, err error)
}

type containerExecuteStructureTestsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newContainerExecuteStructureTestsUtils() containerExecuteStructureTestsUtils {
	utils := containerExecuteStructureTestsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func containerExecuteStructureTests(config containerExecuteStructureTestsOptions, _ *telemetry.CustomData) {
	utils := newContainerExecuteStructureTestsUtils()
	err := runContainerExecuteStructureTests(&config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func findConfigFiles(pattern string, utils containerExecuteStructureTestsUtils) ([]string, error) {
	files, err := utils.Glob(pattern)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func runContainerExecuteStructureTests(config *containerExecuteStructureTestsOptions, utils containerExecuteStructureTestsUtils) error {
	containerStructureTestsExecutable := "./container-structure-test"
	var parameters []string
	parameters = append(parameters, "test")
	configFiles, err := findConfigFiles(config.TestConfiguration, utils)
	if err != nil {
		return errors.Wrapf(err, "failed to find config files, error: %v", err)
	}
	if len(configFiles) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("config files mustn't be missing")
	}
	for _, config := range configFiles {
		parameters = append(parameters, "--config", config)
	}
	if config.TestDriver != "" {
		if config.TestDriver != "docker" && config.TestDriver != "tar" {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("test driver %s is incorrect. Possible drivers: docker, tar", config.TestDriver)
		}
		parameters = append(parameters, "--driver", config.TestDriver)
	} else if os.Getenv("ON_K8S") == "true" {
		parameters = append(parameters, "--driver", "tar")
	}
	if config.PullImage {
		parameters = append(parameters, "--pull")
	}
	parameters = append(parameters, "--image", config.TestImage)
	parameters = append(parameters, "--test-report", config.TestReportFilePath)
	if GeneralConfig.Verbose {
		parameters = append(parameters, "--verbosity", "debug")
	}
	err = utils.RunExecutable(containerStructureTestsExecutable, parameters...)
	if err != nil {
		commandLine := append([]string{containerStructureTestsExecutable}, parameters...)
		return errors.Wrapf(err, "failed to run executable, command: '%s', error: %v", commandLine, err)
	}

	return nil
}
