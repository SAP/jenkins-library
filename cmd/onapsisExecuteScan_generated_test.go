//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnapsisExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := OnapsisExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "onapsisExecuteScan", testCmd.Use, "command name incorrect")

}
