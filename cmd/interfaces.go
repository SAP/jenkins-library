package cmd

import (
	"io"
)

type runner interface {
	SetDir(d string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type execRunner interface {
	runner
	RunExecutable(e string, p ...string) error
}

type envExecRunner interface {
	execRunner
	SetEnv(e []string)
}

type shellRunner interface {
	runner
	RunShell(s string, c string) error
}
