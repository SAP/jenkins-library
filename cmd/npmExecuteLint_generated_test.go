package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmExecuteLintCommand(t *testing.T) {
	t.Parallel()

	testCmd := NpmExecuteLintCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "npmExecuteLint", testCmd.Use, "command name incorrect")

}
