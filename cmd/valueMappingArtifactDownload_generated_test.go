package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueMappingArtifactDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ValueMappingArtifactDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "valueMappingArtifactDownload", testCmd.Use, "command name incorrect")

}
