package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFortifyExecuteScanCommand(t *testing.T) {

	testCmd := FortifyExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "fortifyExecuteScan", testCmd.Use, "command name incorrect")

}
