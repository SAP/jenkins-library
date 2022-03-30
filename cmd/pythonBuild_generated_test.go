package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPythonBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := PythonBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "pythonBuild", testCmd.Use, "command name incorrect")

}
