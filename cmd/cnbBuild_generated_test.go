package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCnbBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := CnbBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "cnbBuild", testCmd.Use, "command name incorrect")

}
