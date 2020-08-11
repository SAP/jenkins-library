package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestExecuteCommand(t *testing.T) {

	testCmd := TestExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "testExecute", testCmd.Use, "command name incorrect")

}
