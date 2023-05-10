//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzureBlobUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := AzureBlobUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "azureBlobUpload", testCmd.Use, "command name incorrect")

}
