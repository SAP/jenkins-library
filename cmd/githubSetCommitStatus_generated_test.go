//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubSetCommitStatusCommand(t *testing.T) {
	t.Parallel()

	testCmd := GithubSetCommitStatusCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "githubSetCommitStatus", testCmd.Use, "command name incorrect")

}
