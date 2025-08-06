package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestReqIDFromGitCommand(t *testing.T) {
	t.Parallel()

	testCmd := TransportRequestReqIDFromGitCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestReqIDFromGit", testCmd.Use, "command name incorrect")

}
