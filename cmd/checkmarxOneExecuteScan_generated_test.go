//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckmarxOneExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := CheckmarxOneExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "checkmarxOneExecuteScan", testCmd.Use, "command name incorrect")

}
