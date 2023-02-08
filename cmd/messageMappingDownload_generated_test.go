package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageMappingDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := MessageMappingDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "messageMappingDownload", testCmd.Use, "command name incorrect")

}
