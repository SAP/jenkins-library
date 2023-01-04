package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactTransportCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactTransportCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactTransport", testCmd.Use, "command name incorrect")

}
