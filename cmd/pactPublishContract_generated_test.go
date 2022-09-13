package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPactPublishContractCommand(t *testing.T) {
	t.Parallel()

	testCmd := PactPublishContractCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "pactPublishContract", testCmd.Use, "command name incorrect")

}
