package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitCheckPVCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapAddonAssemblyKitCheckPVCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitCheckPV", testCmd.Use, "command name incorrect")

}
