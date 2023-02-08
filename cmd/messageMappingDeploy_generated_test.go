package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageMappingDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := MessageMappingDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "messageMappingDeploy", testCmd.Use, "command name incorrect")

}
