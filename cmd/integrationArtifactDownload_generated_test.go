package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactDownload", testCmd.Use, "command name incorrect")
}
