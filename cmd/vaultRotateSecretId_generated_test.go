package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVaultRotateSecretIdCommand(t *testing.T) {
	t.Parallel()

	testCmd := VaultRotateSecretIdCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "vaultRotateSecretId", testCmd.Use, "command name incorrect")
}
