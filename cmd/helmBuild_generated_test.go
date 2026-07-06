//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := HelmBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "helmBuild", testCmd.Use, "command name incorrect")

}
