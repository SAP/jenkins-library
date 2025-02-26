//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildahExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := BuildahExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "buildahExecute", testCmd.Use, "command name incorrect")

}
