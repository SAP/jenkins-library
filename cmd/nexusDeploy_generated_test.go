package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNexusDeployCommand(t *testing.T) {

	testCmd := NexusDeployCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "nexusDeploy", testCmd.Use, "command name incorrect")

}
