package cmd

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func karmaExecuteTests(myKarmaExecuteTestsOptions karmaExecuteTestsOptions) error {
	c := command.Command{}
	// reroute command output to loging framework
	// also log stdout as Karma reports into it
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	telemetry.SendTelemetry(&telemetry.CustomData{})

	runKarma(myKarmaExecuteTestsOptions, &c)
	return nil
}

func runKarma(myKarmaExecuteTestsOptions karmaExecuteTestsOptions, command execRunner) {
	installCommandTokens := tokenize(myKarmaExecuteTestsOptions.InstallCommand)
	command.Dir(myKarmaExecuteTestsOptions.ModulePath)
	err := command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		log.Entry().
			WithError(err).
			WithField("command", myKarmaExecuteTestsOptions.InstallCommand).
			Fatal("failed to execute install command")
	}

	runCommandTokens := tokenize(myKarmaExecuteTestsOptions.RunCommand)
	command.Dir(myKarmaExecuteTestsOptions.ModulePath)
	err = command.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)
	if err != nil {
		log.Entry().
			WithError(err).
			WithField("command", myKarmaExecuteTestsOptions.RunCommand).
			Fatal("failed to execute run command")
	}
}

func tokenize(command string) []string {
	return strings.Split(command, " ")
}
