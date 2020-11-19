package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHadolintExecuteCommand(t *testing.T) {

	testCmd := HadolintExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "hadolintExecute", testCmd.Use, "command name incorrect")

}
