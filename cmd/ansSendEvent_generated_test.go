package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnsSendEventCommand(t *testing.T) {
	t.Parallel()

	testCmd := AnsSendEventCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "ansSendEvent", testCmd.Use, "command name incorrect")

}
