package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckmarxoneExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := CheckmarxoneExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "checkmarxoneExecuteScan", testCmd.Use, "command name incorrect")

}
