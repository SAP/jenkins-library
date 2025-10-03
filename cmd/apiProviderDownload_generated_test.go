package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiProviderDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiProviderDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiProviderDownload", testCmd.Use, "command name incorrect")
}
