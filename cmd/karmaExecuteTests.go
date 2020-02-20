package cmd

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func karmaExecuteTests(config karmaExecuteTestsOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	// reroute command output to loging framework
	// also log stdout as Karma reports into it
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	runKarma(config, &c)
}

func runKarma(config karmaExecuteTestsOptions, command execRunner) {
	installCommandTokens := tokenize(config.InstallCommand)
	command.Dir(config.ModulePath)
	err := command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		log.Entry().
			WithError(err).
			WithField("command", config.InstallCommand).
			Fatal("failed to execute install command")
	}

	runCommandTokens := tokenize(config.RunCommand)
	command.Dir(config.ModulePath)
	err = command.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)
	if err != nil {
		log.Entry().
			WithError(err).
			WithField("command", config.RunCommand).
			Fatal("failed to execute run command")
	}
}

func tokenize(command string) []string {
	return strings.Split(command, " ")
}
