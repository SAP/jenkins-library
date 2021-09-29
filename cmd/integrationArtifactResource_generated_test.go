package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactResourceCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactResourceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactResource", testCmd.Use, "command name incorrect")

}
