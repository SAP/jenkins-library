//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := HelmExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "helmExecute", testCmd.Use, "command name incorrect")

}
