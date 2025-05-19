//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactUpdateConfigurationCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactUpdateConfigurationCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactUpdateConfiguration", testCmd.Use, "command name incorrect")

}
