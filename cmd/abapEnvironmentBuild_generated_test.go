//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentBuild", testCmd.Use, "command name incorrect")

}
