//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapLandscapePortalUpdateAddOnProductCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapLandscapePortalUpdateAddOnProductCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapLandscapePortalUpdateAddOnProduct", testCmd.Use, "command name incorrect")

}
