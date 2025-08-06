package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitCheckCVsCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapAddonAssemblyKitCheckCVsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitCheckCVs", testCmd.Use, "command name incorrect")
}
