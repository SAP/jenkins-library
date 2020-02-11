package cmd

import (
	"io"
	"os"
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

type fileUtils interface {
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
}
