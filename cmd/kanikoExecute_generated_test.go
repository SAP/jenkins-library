//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKanikoExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := KanikoExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "kanikoExecute", testCmd.Use, "command name incorrect")

}
