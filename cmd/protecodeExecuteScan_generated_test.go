package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtecodeExecuteScanCommand(t *testing.T) {

	testCmd := ProtecodeExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "protecodeExecuteScan", testCmd.Use, "command name incorrect")

}
