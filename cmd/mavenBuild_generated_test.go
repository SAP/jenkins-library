package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenBuildCommand(t *testing.T) {

	testCmd := MavenBuildCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "mavenBuild", testCmd.Use, "command name incorrect")

}
