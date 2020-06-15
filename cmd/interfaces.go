package cmd

import (
	"io"
)

type runner interface {
	SetDir(d string)
	SetEnv(e []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type execRunnerExecution interface {
	Kill() error
	Wait() error
}

type execRunner interface {
	runner
	RunExecutable(e string, p ...string) error
	RunExecutableInBackground(executable string, params ...string) (*execRunnerExecution, error)
}

type shellRunner interface {
	runner
	RunShell(s string, c string) error
}
