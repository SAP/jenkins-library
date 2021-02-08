package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestCreateSOLMANCommand(t *testing.T) {
	t.Parallel()

	testCmd := TransportRequestCreateSOLMANCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestCreateSOLMAN", testCmd.Use, "command name incorrect")

}
