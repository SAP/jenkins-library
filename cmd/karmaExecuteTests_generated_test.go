package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKarmaExecuteTestsCommand(t *testing.T) {

	testCmd := KarmaExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "karmaExecuteTests", testCmd.Use, "command name incorrect")

}
