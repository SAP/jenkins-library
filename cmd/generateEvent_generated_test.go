//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateEventCommand(t *testing.T) {
	t.Parallel()

	testCmd := GenerateEventCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "generateEvent", testCmd.Use, "command name incorrect")

}
