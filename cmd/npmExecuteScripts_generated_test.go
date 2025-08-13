package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmExecuteScriptsCommand(t *testing.T) {
	t.Parallel()

	testCmd := NpmExecuteScriptsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "npmExecuteScripts", testCmd.Use, "command name incorrect")
}
