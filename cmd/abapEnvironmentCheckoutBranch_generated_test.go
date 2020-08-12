package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentCheckoutBranchCommand(t *testing.T) {

	testCmd := AbapEnvironmentCheckoutBranchCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "abapEnvironmentCheckoutBranch", testCmd.Use, "command name incorrect")

}
