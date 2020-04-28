package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsDeployCommitCommand(t *testing.T) {

	testCmd := GctsDeployCommitCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "gctsDeployCommit", testCmd.Use, "command name incorrect")

}
