package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentRunATCCheckCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentRunATCCheckCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentRunATCCheck", testCmd.Use, "command name incorrect")
}
