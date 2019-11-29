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

type shellRunner interface {
	RunShell(s string, c string) error
	Dir(d string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type fileUtils interface {
	// In case "path" is relative it is evaluated relative to the
	// current working directory.
	FileExists(path string) (bool, error)

	// Copies a file from source to test (currently intended for files only, not folders)
	CopyFile(source string, dest string) (int64, error)

	RemoveFile(path string) error

	OpenFile(path string) (*os.File, error)
}
