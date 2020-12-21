package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteNewmanCommand(t *testing.T) {
	t.Parallel()

	testCmd := ExecuteNewmanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "executeNewman", testCmd.Use, "command name incorrect")

}
