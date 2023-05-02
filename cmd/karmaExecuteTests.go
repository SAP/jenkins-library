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
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())
	runKarma(config, &c)
}

func runKarma(config karmaExecuteTestsOptions, command command.ExecRunner) {
	installCommandTokens := tokenize(config.InstallCommand)
	runCommandTokens := tokenize(config.RunCommand)
	modulePaths := config.Modules

	if GeneralConfig.Verbose {
		runCommandTokens = append(runCommandTokens, "--", "--log-level", "DEBUG")
	}

	for _, module := range modulePaths {
		command.SetDir(module)
		err := command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
		if err != nil {
			log.SetErrorCategory(log.ErrorCustom)
			log.Entry().
				WithError(err).
				WithField("command", config.InstallCommand).
				Fatal("failed to execute install command")
		}

		command.SetDir(module)
		err = command.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)
		if err != nil {
			log.SetErrorCategory(log.ErrorTest)
			log.Entry().
				WithError(err).
				WithField("command", config.RunCommand).
				Fatal("failed to execute run command")
		}
	}
}

func tokenize(command string) []string {
	return strings.Split(command, " ")
}
