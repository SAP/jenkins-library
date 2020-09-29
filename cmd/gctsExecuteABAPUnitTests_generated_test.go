package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsExecuteABAPUnitTestsCommand(t *testing.T) {

	testCmd := GctsExecuteABAPUnitTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gctsExecuteABAPUnitTests", testCmd.Use, "command name incorrect")

}
