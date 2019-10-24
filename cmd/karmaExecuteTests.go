package cmd

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/pkg/errors"
)

type execRunner interface {
	RunExecutable(e string, p ...string) error
	Dir(d string)
}

func karmaExecuteTests(myKarmaExecuteTestsOptions karmaExecuteTestsOptions) error {
	c := command.Command{}
	return runKarma(myKarmaExecuteTestsOptions, &c)
}

func runKarma(myKarmaExecuteTestsOptions karmaExecuteTestsOptions, command execRunner) error {
	installCommandTokens := tokenize(myKarmaExecuteTestsOptions.InstallCommand)
	command.Dir(myKarmaExecuteTestsOptions.ModulePath)
	err := command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute install command '%v'", myKarmaExecuteTestsOptions.InstallCommand)
	}

	runCommandTokens := tokenize(myKarmaExecuteTestsOptions.RunCommand)
	command.Dir(myKarmaExecuteTestsOptions.ModulePath)
	err = command.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute run command '%v'", myKarmaExecuteTestsOptions.RunCommand)
	}

	return nil
}

func tokenize(command string) []string {
	return strings.Split(command, " ")
}
