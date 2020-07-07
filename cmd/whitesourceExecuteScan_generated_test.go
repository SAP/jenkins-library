package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhitesourceExecuteScanCommand(t *testing.T) {

	testCmd := WhitesourceExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "whitesourceExecuteScan", testCmd.Use, "command name incorrect")

}
