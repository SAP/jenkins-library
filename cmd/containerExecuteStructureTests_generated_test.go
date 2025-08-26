package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerExecuteStructureTestsCommand(t *testing.T) {
	t.Parallel()

	testCmd := ContainerExecuteStructureTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "containerExecuteStructureTests", testCmd.Use, "command name incorrect")
}
