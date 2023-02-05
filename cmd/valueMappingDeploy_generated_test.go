package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueMappingDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := ValueMappingDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "valueMappingDeploy", testCmd.Use, "command name incorrect")

}
