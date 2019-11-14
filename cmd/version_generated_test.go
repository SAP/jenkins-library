package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {

	testCmd := VersionCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "version", testCmd.Use, "command name incorrect")

}
