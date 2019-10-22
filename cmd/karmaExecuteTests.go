package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/SAP/jenkins-library/pkg/command"
)

func karmaExecuteTests(myKarmaExecuteTestsOptions karmaExecuteTestsOptions) error {

	installCommandTokens := tokenize(myKarmaExecuteTestsOptions.InstallCommand)
	s := command.Executable{
		Dir:        myKarmaExecuteTestsOptions.ModulePath,
		Executable: installCommandTokens[0],
		Parameters: installCommandTokens[1:],
	}
	err := s.Run()
	if err != nil {
		return errors.Wrapf(err, "failed to execute install command '%v'", myKarmaExecuteTestsOptions.InstallCommand)
	}

	runCommandTokens := tokenize(myKarmaExecuteTestsOptions.RunCommand)
	s = command.Executable{
		Dir:        myKarmaExecuteTestsOptions.ModulePath,
		Executable: runCommandTokens[0],
		Parameters: runCommandTokens[1:],
	}
	err = s.Run()
	if err != nil {
		return errors.Wrapf(err, "failed to execute run command '%v'", myKarmaExecuteTestsOptions.RunCommand)
	}

	return nil
}

func tokenize(command string) []string {
	return strings.Split(command, " ")
}


