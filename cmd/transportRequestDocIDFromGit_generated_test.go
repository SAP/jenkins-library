//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestDocIDFromGitCommand(t *testing.T) {
	t.Parallel()

	testCmd := TransportRequestDocIDFromGitCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestDocIDFromGit", testCmd.Use, "command name incorrect")

}
