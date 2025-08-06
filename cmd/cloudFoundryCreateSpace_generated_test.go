package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateSpaceCommand(t *testing.T) {
	t.Parallel()

	testCmd := CloudFoundryCreateSpaceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "cloudFoundryCreateSpace", testCmd.Use, "command name incorrect")
}
