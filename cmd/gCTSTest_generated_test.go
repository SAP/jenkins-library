package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCTSTestCommand(t *testing.T) {
	t.Parallel()

	testCmd := GCTSTestCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gCTSTest", testCmd.Use, "command name incorrect")

}
