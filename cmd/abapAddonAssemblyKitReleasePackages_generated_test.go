package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitReleasePackagesCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapAddonAssemblyKitReleasePackagesCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitReleasePackages", testCmd.Use, "command name incorrect")
}
