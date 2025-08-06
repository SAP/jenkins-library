package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactUnDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactUnDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactUnDeploy", testCmd.Use, "command name incorrect")
}
