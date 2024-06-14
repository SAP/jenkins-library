//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContrastExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := ContrastExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "contrastExecuteScan", testCmd.Use, "command name incorrect")

}
