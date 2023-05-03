//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubPublishReleaseCommand(t *testing.T) {
	t.Parallel()

	testCmd := GithubPublishReleaseCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "githubPublishRelease", testCmd.Use, "command name incorrect")

}
