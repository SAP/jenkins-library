package cmd

import (
	"io"
	"os/exec"
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
	RunExecutableInBackground(executable string, params ...string) (*exec.Cmd, error)
}

type shellRunner interface {
	runner
	RunShell(s string, c string) error
}
