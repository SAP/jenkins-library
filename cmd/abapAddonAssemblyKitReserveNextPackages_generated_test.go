package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitReserveNextPackagesCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapAddonAssemblyKitReserveNextPackagesCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitReserveNextPackages", testCmd.Use, "command name incorrect")
}
