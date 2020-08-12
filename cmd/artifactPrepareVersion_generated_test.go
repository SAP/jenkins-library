package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArtifactPrepareVersionCommand(t *testing.T) {

	testCmd := ArtifactPrepareVersionCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "artifactPrepareVersion", testCmd.Use, "command name incorrect")

}
