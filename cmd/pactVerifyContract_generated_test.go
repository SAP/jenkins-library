package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPactVerifyContractCommand(t *testing.T) {
	t.Parallel()

	testCmd := PactVerifyContractCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "pactVerifyContract", testCmd.Use, "command name incorrect")

}
