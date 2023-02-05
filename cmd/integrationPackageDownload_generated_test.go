package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationPackageDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationPackageDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationPackageDownload", testCmd.Use, "command name incorrect")

}
