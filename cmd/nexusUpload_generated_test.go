package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNexusUploadCommand(t *testing.T) {

	testCmd := NexusUploadCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "nexusUpload", testCmd.Use, "command name incorrect")

}
