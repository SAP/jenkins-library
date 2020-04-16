package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeJsBuildCommand(t *testing.T) {

	testCmd := NodeJsBuildCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "nodeJsBuild", testCmd.Use, "command name incorrect")

}
