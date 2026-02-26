//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := GctsDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gctsDeploy", testCmd.Use, "command name incorrect")

}
