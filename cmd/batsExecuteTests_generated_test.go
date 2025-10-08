package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatsExecuteTestsCommand(t *testing.T) {
	t.Parallel()

	testCmd := BatsExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "batsExecuteTests", testCmd.Use, "command name incorrect")
}
