package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckmarxExecuteScanCommand(t *testing.T) {

	testCmd := CheckmarxExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "checkmarxExecuteScan", testCmd.Use, "command name incorrect")

}
