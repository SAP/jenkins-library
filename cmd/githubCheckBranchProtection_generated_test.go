package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubCheckBranchProtectionCommand(t *testing.T) {
	t.Parallel()

	testCmd := GithubCheckBranchProtectionCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "githubCheckBranchProtection", testCmd.Use, "command name incorrect")

}
