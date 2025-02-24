//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildkitExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := BuildkitExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "buildkitExecute", testCmd.Use, "command name incorrect")

}
