package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKanikoExecuteCommand(t *testing.T) {

	testCmd := KanikoExecuteCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "kanikoExecute", testCmd.Use, "command name incorrect")

}
