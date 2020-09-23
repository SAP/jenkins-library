package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitRegisterPackagesCommand(t *testing.T) {

	testCmd := AbapAddonAssemblyKitRegisterPackagesCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitRegisterPackages", testCmd.Use, "command name incorrect")

}
