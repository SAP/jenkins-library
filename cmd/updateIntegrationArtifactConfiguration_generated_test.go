package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateIntegrationArtifactConfigurationCommand(t *testing.T) {
	t.Parallel()

	testCmd := UpdateIntegrationArtifactConfigurationCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "updateIntegrationArtifactConfiguration", testCmd.Use, "command name incorrect")

}
