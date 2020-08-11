package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubCreatePullRequestCommand(t *testing.T) {

	testCmd := GithubCreatePullRequestCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "githubCreatePullRequest", testCmd.Use, "command name incorrect")

}
