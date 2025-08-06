package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiProviderUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiProviderUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiProviderUpload", testCmd.Use, "command name incorrect")

}
