//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := DetectExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "detectExecuteScan", testCmd.Use, "command name incorrect")

}
