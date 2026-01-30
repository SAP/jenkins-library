//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelloExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := HelloExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "helloExecute", testCmd.Use, "command name incorrect")

}
