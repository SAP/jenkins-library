package cmd

import (
	"io"
)

type Runner interface {
	Dir(d string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type ExecRunner interface {
	Runner
	RunExecutable(e string, p ...string) error
}

type EnvExecRunner interface {
	ExecRunner
	Env(e []string)
}

type ShellRunner interface {
	Runner
	RunShell(s string, c string) error
}
