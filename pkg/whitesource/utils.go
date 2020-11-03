package whitesource

import (
	"io"
	"net/http"
	"os"
)

// File defines the method subset we use from os.File
type File interface {
	io.Writer
	io.StringWriter
	io.Closer
}

// Utils captures all external functionality that needs to be exchangeable in tests.
type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(executable string, params ...string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error

	Chdir(path string) error
	Getwd() (string, error)
	MkdirAll(path string, perm os.FileMode) error
	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
	FileRename(oldPath, newPath string) error
	RemoveAll(path string) error
	FileOpen(name string, flag int, perm os.FileMode) (File, error)

	FindPackageJSONFiles(config *ScanOptions) ([]string, error)
	InstallAllNPMDependencies(config *ScanOptions, packageJSONFiles []string) error
}
