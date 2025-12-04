//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtecodeExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := ProtecodeExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "protecodeExecuteScan", testCmd.Use, "command name incorrect")

}
