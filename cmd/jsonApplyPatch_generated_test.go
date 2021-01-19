package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonApplyPatchCommand(t *testing.T) {
	t.Parallel()

	testCmd := JsonApplyPatchCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "jsonApplyPatch", testCmd.Use, "command name incorrect")

}
