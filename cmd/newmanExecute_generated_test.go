package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewmanExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := NewmanExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "newmanExecute", testCmd.Use, "command name incorrect")

}
