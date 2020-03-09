package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenStaticCodeChecksCommand(t *testing.T) {

	testCmd := MavenStaticCodeChecksCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "mavenStaticCodeChecks", testCmd.Use, "command name incorrect")

}
