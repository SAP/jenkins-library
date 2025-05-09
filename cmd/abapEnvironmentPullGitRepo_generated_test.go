//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentPullGitRepoCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentPullGitRepoCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentPullGitRepo", testCmd.Use, "command name incorrect")

}
