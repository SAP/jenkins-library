//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := TerraformExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "terraformExecute", testCmd.Use, "command name incorrect")

}
