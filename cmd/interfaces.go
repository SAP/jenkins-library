package cmd

import (
	"io"
)

type execRunner interface {
	RunExecutable(e string, p ...string) error
	Dir(d string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type envExecRunner interface {
	execRunner
	Env(e []string)
}

type shellRunner interface {
	RunShell(s string, c string) error
	Dir(d string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}
