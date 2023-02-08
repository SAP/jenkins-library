package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageMappingUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := MessageMappingUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "messageMappingUpload", testCmd.Use, "command name incorrect")

}
