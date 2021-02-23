package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenExecuteIntegrationCommand(t *testing.T) {
	t.Parallel()

	testCmd := MavenExecuteIntegrationCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "mavenExecuteIntegration", testCmd.Use, "command name incorrect")

}
