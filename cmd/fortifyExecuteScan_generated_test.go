package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFortifyExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := FortifyExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "fortifyExecuteScan", testCmd.Use, "command name incorrect")
}
