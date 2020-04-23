package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmExecuteScriptsCommand(t *testing.T) {

	testCmd := NpmExecuteScriptsCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "npmExecuteScripts", testCmd.Use, "command name incorrect")

}
