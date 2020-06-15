package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"io"
)

type runner interface {
	SetDir(d string)
	SetEnv(e []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type execRunner interface {
	runner
	RunExecutable(e string, p ...string) error
	RunExecutableInBackground(executable string, params ...string) (command.Execution, error)
}

type shellRunner interface {
	runner
	RunShell(s string, c string) error
}
