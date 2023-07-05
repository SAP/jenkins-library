//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitRegisterPackagesCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapAddonAssemblyKitRegisterPackagesCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitRegisterPackages", testCmd.Use, "command name incorrect")

}
