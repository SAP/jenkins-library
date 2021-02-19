package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubCommentIssueCommand(t *testing.T) {
	t.Parallel()

	testCmd := GithubCommentIssueCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "githubCommentIssue", testCmd.Use, "command name incorrect")

}
