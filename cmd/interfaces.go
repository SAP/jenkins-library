package cmd

import (
	"io"
)

type runner interface {
	Dir(d string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type execRunner interface {
	runner
	RunExecutable(e string, p ...string) error
}

type envExecRunner interface {
	execRunner
	Env(e []string)
}

type shellRunner interface {
	runner
	RunShell(s string, c string) error
}
