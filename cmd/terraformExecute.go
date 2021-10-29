package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type terraformExecuteUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
}

type terraformExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newTerraformExecuteUtils() terraformExecuteUtils {
	utils := terraformExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func terraformExecute(config terraformExecuteOptions, telemetryData *telemetry.CustomData) {
	utils := newTerraformExecuteUtils()

	err := runTerraformExecute(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTerraformExecute(config *terraformExecuteOptions, telemetryData *telemetry.CustomData, utils terraformExecuteUtils) error {
	args := []string{}

	if config.Command == "apply" {
		args = append(args, "-auto-approve")
	}

	if (config.Command == "apply" || config.Command == "plan") && config.TerraformSecrets != "" {
		args = append(args, fmt.Sprintf("-var-file=%s", config.TerraformSecrets))
	}

	if config.AdditionalArgs != nil {
		args = append(args, config.AdditionalArgs...)
	}

	if config.Init {
		err := runTerraform(utils, "init", []string{}, config.GlobalOptions)

		if err != nil {
			return err
		}
	}

	return runTerraform(utils, config.Command, args, config.GlobalOptions)
}

func runTerraform(utils terraformExecuteUtils, command string, args []string, globalOptions []string) error {
	cliArgs := []string{}

	if globalOptions != nil {
		cliArgs = append(cliArgs, globalOptions...)
	}

	cliArgs = append(cliArgs, command)

	if args != nil {
		cliArgs = append(cliArgs, args...)
	}

	return utils.RunExecutable("terraform", cliArgs...)
}
