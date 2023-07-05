//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiKeyValueMapDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiKeyValueMapDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiKeyValueMapDownload", testCmd.Use, "command name incorrect")

}
