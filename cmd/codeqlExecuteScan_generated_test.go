//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeqlExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := CodeqlExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "codeqlExecuteScan", testCmd.Use, "command name incorrect")

}
