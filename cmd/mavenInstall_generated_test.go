package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenInstallCommand(t *testing.T) {

	testCmd := MavenInstallCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "mavenInstall", testCmd.Use, "command name incorrect")

}
