package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmExecuteTestsCommand(t *testing.T) {
	t.Parallel()

	testCmd := NpmExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "npmExecuteTests", testCmd.Use, "command name incorrect")
}
