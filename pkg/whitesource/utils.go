package whitesource

import (
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/maven"
)

// File defines the method subset we use from os.File
type File interface {
	io.Writer
	io.StringWriter
	io.Closer
}

// Utils captures all external functionality that needs to be exchangeable in tests.
type Utils interface {
	maven.Utils

	Chdir(path string) error
	Getwd() (string, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
	FileRename(oldPath, newPath string) error
	GetExitCode() int
	RemoveAll(path string) error
	FileOpen(name string, flag int, perm os.FileMode) (File, error)

	FindPackageJSONFiles(config *ScanOptions) ([]string, error)
	InstallAllNPMDependencies(config *ScanOptions, packageJSONFiles []string) error
}
