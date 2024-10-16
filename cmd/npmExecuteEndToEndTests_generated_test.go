//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmExecuteEndToEndTestsCommand(t *testing.T) {
	t.Parallel()

	testCmd := NpmExecuteEndToEndTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "npmExecuteEndToEndTests", testCmd.Use, "command name incorrect")

}
