package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNexusUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := NexusUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "nexusUpload", testCmd.Use, "command name incorrect")

}
