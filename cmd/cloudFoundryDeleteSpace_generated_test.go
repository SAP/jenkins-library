//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeleteSpaceCommand(t *testing.T) {
	t.Parallel()

	testCmd := CloudFoundryDeleteSpaceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "cloudFoundryDeleteSpace", testCmd.Use, "command name incorrect")

}
