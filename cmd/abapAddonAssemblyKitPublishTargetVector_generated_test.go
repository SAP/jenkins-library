//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapAddonAssemblyKitPublishTargetVectorCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapAddonAssemblyKitPublishTargetVectorCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapAddonAssemblyKitPublishTargetVector", testCmd.Use, "command name incorrect")

}
