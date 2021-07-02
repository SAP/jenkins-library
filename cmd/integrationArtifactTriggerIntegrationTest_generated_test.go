package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactTriggerIntegrationTestCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactTriggerIntegrationTestCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactTriggerIntegrationTest", testCmd.Use, "command name incorrect")

}
