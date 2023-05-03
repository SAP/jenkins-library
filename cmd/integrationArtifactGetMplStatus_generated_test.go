//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactGetMplStatusCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactGetMplStatusCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactGetMplStatus", testCmd.Use, "command name incorrect")

}
