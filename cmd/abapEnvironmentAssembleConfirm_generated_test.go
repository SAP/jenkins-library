package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentAssembleConfirmCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentAssembleConfirmCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentAssembleConfirm", testCmd.Use, "command name incorrect")

}
