package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsRollbackCommand(t *testing.T) {
	t.Parallel()

	testCmd := GctsRollbackCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gctsRollback", testCmd.Use, "command name incorrect")

}
