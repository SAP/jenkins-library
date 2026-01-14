//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialdiggerScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := CredentialdiggerScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "credentialdiggerScan", testCmd.Use, "command name incorrect")

}
