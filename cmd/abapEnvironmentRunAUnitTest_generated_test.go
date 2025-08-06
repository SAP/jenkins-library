package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentRunAUnitTestCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentRunAUnitTestCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentRunAUnitTest", testCmd.Use, "command name incorrect")
}
