//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubCreateIssueCommand(t *testing.T) {
	t.Parallel()

	testCmd := GithubCreateIssueCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "githubCreateIssue", testCmd.Use, "command name incorrect")

}
