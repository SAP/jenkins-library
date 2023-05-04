//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsExecuteABAPQualityChecksCommand(t *testing.T) {
	t.Parallel()

	testCmd := GctsExecuteABAPQualityChecksCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gctsExecuteABAPQualityChecks", testCmd.Use, "command name incorrect")

}
