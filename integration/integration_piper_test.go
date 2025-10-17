//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPiperIntegration ./integration/...

package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
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
