package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenExecuteCommand(t *testing.T) {

	testCmd := MavenExecuteCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "mavenExecute", testCmd.Use, "command name incorrect")

}
