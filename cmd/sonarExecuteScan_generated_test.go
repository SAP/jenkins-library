//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSonarExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := SonarExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sonarExecuteScan", testCmd.Use, "command name incorrect")

}
