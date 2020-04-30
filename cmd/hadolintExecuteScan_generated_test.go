package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHadolintExecuteScanCommand(t *testing.T) {

	testCmd := HadolintExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "hadolintExecuteScan", testCmd.Use, "command name incorrect")

}
