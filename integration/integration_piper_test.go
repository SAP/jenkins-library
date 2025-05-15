//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPiperIntegration ./integration/...

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

func TestPiperIntegrationHelp(t *testing.T) {
	// t.Parallel()
	piperHelpCmd := command.Command{}

	var commandOutput bytes.Buffer
	piperHelpCmd.Stdout(&commandOutput)

	err := piperHelpCmd.RunExecutable(getPiperExecutable(), "--help")

	assert.NoError(t, err, "Calling piper --help failed")
	assert.Contains(t, commandOutput.String(), "Use \"piper [command] --help\" for more information about a command.")
}

func getPiperExecutable() string {
	if p := os.Getenv("PIPER_INTEGRATION_EXECUTABLE"); len(p) > 0 {
		fmt.Println("Piper executable for integration test: " + p)
		return p
	}

	f := piperutils.Files{}
	wd, _ := os.Getwd()
	localPiper := path.Join(wd, "..", "piper")
	exists, _ := f.FileExists(localPiper)
	if exists {
		fmt.Println("Piper executable for integration test: " + localPiper)
		return localPiper
	}

	fmt.Println("Piper executable for integration test: Using 'piper' from PATH")
	return "piper"
}

// copyDir copies a directory
func copyDir(source string, target string) error {
	var err error
	var fileInfo []os.DirEntry
	var sourceInfo os.FileInfo

	if sourceInfo, err = os.Stat(source); err != nil {
		return err
	}

	if err = os.MkdirAll(target, sourceInfo.Mode()); err != nil {
		return err
	}

	if fileInfo, err = os.ReadDir(source); err != nil {
		return err
	}
	for _, info := range fileInfo {
		sourcePath := path.Join(source, info.Name())
		targetPath := path.Join(target, info.Name())

		if info.IsDir() {
			if err = copyDir(sourcePath, targetPath); err != nil {
				return err
			}
		} else {
			if err = copyFile(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(source, target string) error {
	var err error
	var sourceFile *os.File
	var targetFile *os.File
	var sourceInfo os.FileInfo

	if sourceFile, err = os.Open(source); err != nil {
		return err
	}
	defer sourceFile.Close()

	if targetFile, err = os.Create(target); err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err = io.Copy(targetFile, sourceFile); err != nil {
		return err
	}
	if sourceInfo, err = os.Stat(source); err != nil {
		return err
	}
	return os.Chmod(target, sourceInfo.Mode())
}

// createTmpDir calls t.TempDir() and returns the path name after the evaluation
// of any symbolic links.
//
// On Docker for Mac, t.TempDir() returns e.g.
// /var/folders/bl/wbxjgtzx7j5_mjsmfr3ynlc00000gp/T/<the-test-name>/001
func createTmpDir(t *testing.T) (string, error) {
	tmpDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}
