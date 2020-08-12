package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenExecuteStaticCodeChecksCommand(t *testing.T) {

	testCmd := MavenExecuteStaticCodeChecksCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "mavenExecuteStaticCodeChecks", testCmd.Use, "command name incorrect")

}
