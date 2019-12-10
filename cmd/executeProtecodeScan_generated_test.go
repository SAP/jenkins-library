package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteProtecodeScanCommand(t *testing.T) {

	testCmd := ExecuteProtecodeScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "executeProtecodeScan", testCmd.Use, "command name incorrect")

}
