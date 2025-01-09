//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubSecretScanningReportCommand(t *testing.T) {
	t.Parallel()

	testCmd := GithubSecretScanningReportCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "githubSecretScanningReport", testCmd.Use, "command name incorrect")

}
