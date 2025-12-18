//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerSaveImageCommand(t *testing.T) {
	t.Parallel()

	testCmd := ContainerSaveImageCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "containerSaveImage", testCmd.Use, "command name incorrect")

}
