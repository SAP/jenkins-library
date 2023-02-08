package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationPackageUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationPackageUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationPackageUpload", testCmd.Use, "command name incorrect")

}
