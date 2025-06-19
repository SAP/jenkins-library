//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentCheckoutBranchCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentCheckoutBranchCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentCheckoutBranch", testCmd.Use, "command name incorrect")

}
