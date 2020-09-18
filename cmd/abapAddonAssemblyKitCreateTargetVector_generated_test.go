package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitCreateTargetVectorCommand(t *testing.T) {

	testCmd := AbapAddonAssemblyKitCreateTargetVectorCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitCreateTargetVector", testCmd.Use, "command name incorrect")

}
