package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTmsExportCommand(t *testing.T) {
	t.Parallel()

	testCmd := TmsExportCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "tmsExport", testCmd.Use, "command name incorrect")
}
