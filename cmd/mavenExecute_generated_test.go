//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := MavenExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "mavenExecute", testCmd.Use, "command name incorrect")

}
