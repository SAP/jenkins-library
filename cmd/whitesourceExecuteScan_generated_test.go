//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhitesourceExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := WhitesourceExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "whitesourceExecuteScan", testCmd.Use, "command name incorrect")

}
