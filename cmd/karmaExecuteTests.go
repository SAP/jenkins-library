package cmd

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"

	// TODO move dependency to pkg/log
	"github.com/sirupsen/logrus"
)

func karmaExecuteTests(myKarmaExecuteTestsOptions karmaExecuteTestsOptions) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout = log.Logger().Writer()
	c.Stderr = log.Logger().WriterLevel(logrus.ErrorLevel)
	return runKarma(myKarmaExecuteTestsOptions, &c)
}

func runKarma(myKarmaExecuteTestsOptions karmaExecuteTestsOptions, command execRunner) error {
	installCommandTokens := tokenize(myKarmaExecuteTestsOptions.InstallCommand)
	command.Dir(myKarmaExecuteTestsOptions.ModulePath)
	err := command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		log.Logger().
			WithError(err).
			WithField("command", myKarmaExecuteTestsOptions.InstallCommand).
			Fatal("failed to execute install command")
	}

	runCommandTokens := tokenize(myKarmaExecuteTestsOptions.RunCommand)
	command.Dir(myKarmaExecuteTestsOptions.ModulePath)
	err = command.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)
	if err != nil {
		log.Logger().
			WithError(err).
			WithField("command", myKarmaExecuteTestsOptions.RunCommand).
			Fatal("failed to execute run command")
	}

	return nil
}

func tokenize(command string) []string {
	return strings.Split(command, " ")
}
