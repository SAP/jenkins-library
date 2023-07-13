//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactUpload", testCmd.Use, "command name incorrect")

}
