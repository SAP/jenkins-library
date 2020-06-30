package command

import (
	"io"
)

type runner interface {
	SetDir(d string)
	SetEnv(e []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

// ExecRunner mock for intercepting calls to executables
type ExecRunner interface {
	runner
	RunExecutable(e string, p ...string) error
	RunExecutableInBackground(executable string, params ...string) (Execution, error)
}

// ShellRunner mock for intercepting shell calls
type ShellRunner interface {
	runner
	RunShell(s string, c string) error
}
