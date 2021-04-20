package cmd

import (
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type gaugeExecuteTestsUtils interface {
	FileExists(filename string) (bool, error)
	SetEnv([]string)
	RunShell(shell string, command string) error
}

type gaugeExecuteTestsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newGaugeExecuteTestsUtils() gaugeExecuteTestsUtils {
	utils := gaugeExecuteTestsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func gaugeExecuteTests(config gaugeExecuteTestsOptions, telemetryData *telemetry.CustomData, influx *gaugeExecuteTestsInflux) {
	utils := newGaugeExecuteTestsUtils()

	influx.step_data.fields.gauge = false
	err := runGaugeExecuteTests(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	influx.step_data.fields.gauge = true
}

func runGaugeExecuteTests(config *gaugeExecuteTestsOptions, telemetryData *telemetry.CustomData, utils gaugeExecuteTestsUtils) error {
	var gaugeScript string
	if config.InstallCommand != "" {
		gaugeScript += `export HOME=${HOME:-$(pwd)}
if [ "$HOME" = "/" ]; then export HOME=$(pwd); fi
export PATH=$HOME/bin/gauge:$PATH
mkdir -p $HOME/bin/gauge
` + config.InstallCommand + `
gauge install html-report
gauge install xml-report
`
	}
	if config.LanguageRunner != "" {
		gaugeScript += "gauge install " + config.LanguageRunner + "\n"
	}

	gaugeScript += config.RunCommand

	if config.TestOptions != "" {
		gaugeScript += " " + config.TestOptions
	}

	homedir, _ := os.UserHomeDir()
	path := "PATH=" + os.Getenv("PATH") + ":" + homedir + "/.npm-global/bin"

	utils.SetEnv([]string{path})

	err := utils.RunShell("/bin/bash", gaugeScript)
	if err != nil {
		return fmt.Errorf("failed to run gauge: %w", err)
	}
	return nil
}
