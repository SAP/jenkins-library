package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiKeyValueMapUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiKeyValueMapUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiKeyValueMapUpload", testCmd.Use, "command name incorrect")
}
