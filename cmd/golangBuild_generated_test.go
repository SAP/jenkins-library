//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGolangBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := GolangBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "golangBuild", testCmd.Use, "command name incorrect")

}
