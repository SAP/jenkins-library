//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArtifactPrepareVersionCommand(t *testing.T) {
	t.Parallel()

	testCmd := ArtifactPrepareVersionCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "artifactPrepareVersion", testCmd.Use, "command name incorrect")

}
