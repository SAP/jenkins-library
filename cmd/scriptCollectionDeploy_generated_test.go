package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScriptCollectionDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := ScriptCollectionDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "scriptCollectionDeploy", testCmd.Use, "command name incorrect")

}
