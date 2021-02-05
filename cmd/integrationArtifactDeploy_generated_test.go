package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactDeploy", testCmd.Use, "command name incorrect")

}
