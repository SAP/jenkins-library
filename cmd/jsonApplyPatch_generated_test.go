package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonApplyPatchCommand(t *testing.T) {

	testCmd := JsonApplyPatchCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "jsonApplyPatch", testCmd.Use, "command name incorrect")

}
