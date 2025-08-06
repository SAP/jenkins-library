package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentCloneGitRepoCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentCloneGitRepoCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentCloneGitRepo", testCmd.Use, "command name incorrect")
}
