package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestCreateCTSCommand(t *testing.T) {
	t.Parallel()

	testCmd := TransportRequestCreateCTSCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestCreateCTS", testCmd.Use, "command name incorrect")

}
