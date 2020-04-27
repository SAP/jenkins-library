// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNpmExecuteScripts(t *testing.T) {
	cmd := command.Command{}
	cmd.SetDir("testdata/TestNpmIntegration")

	piperOptions := []string{
		"npmExecuteScripts",
		"--install",
		"--runScripts=ci-build,ci-backend-unit-test",
	}

	var commandOutput bytes.Buffer
	cmd.Stdout(&commandOutput)
	cmd.Stderr(&commandOutput)

	err := cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)
	assert.Contains(t, commandOutput.String(), "Discovered pre-configured npm registry https://example.com")
}
