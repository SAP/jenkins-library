package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeployIntegrationArtifactCommand(t *testing.T) {

	testCmd := DeployIntegrationArtifactCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "deployIntegrationArtifact", testCmd.Use, "command name incorrect")

}
