package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectExecuteScanCommand(t *testing.T) {

	testCmd := DetectExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "detectExecuteScan", testCmd.Use, "command name incorrect")

}
