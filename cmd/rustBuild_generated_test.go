//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRustBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := RustBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "rustBuild", testCmd.Use, "command name incorrect")

}
