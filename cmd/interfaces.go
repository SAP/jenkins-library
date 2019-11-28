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

type shellRunner interface {
	RunShell(s string, c string) error
	Dir(d string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type fileSystem interface {
	// In case "path" is relative it is evaluated relative to the
	// current working directory.
	FileExists(path string) (bool, error)

	// Copies a file from source to test (files only, currently no folder copy)
	Copy(source string, dest string) (int64, error)

	Remove(path string) error
}
