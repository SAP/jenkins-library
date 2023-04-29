//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenExecuteStaticCodeChecksCommand(t *testing.T) {
	t.Parallel()

	testCmd := MavenExecuteStaticCodeChecksCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "mavenExecuteStaticCodeChecks", testCmd.Use, "command name incorrect")

}
