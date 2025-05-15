//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitCheckCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapAddonAssemblyKitCheckCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitCheck", testCmd.Use, "command name incorrect")

}
