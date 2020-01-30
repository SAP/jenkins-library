package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestPiperHelp(t *testing.T) {
	piperHelpCmd := command.Command{}

	var commandOutput bytes.Buffer
	piperHelpCmd.Stdout(&commandOutput)

	err := piperHelpCmd.RunExecutable(getPiperExecutable(), "--help")

	assert.NoError(t, err, "Calling piper --help failed")
	assert.Contains(t, commandOutput.String(), "Use \"piper [command] --help\" for more information about a command.")
}

func getPiperExecutable() string {
	if p := os.Getenv("PIPER_INTEGRATION_EXECUTABLE"); len(p) > 0 {
		return p
	}
	return "piper"
}
