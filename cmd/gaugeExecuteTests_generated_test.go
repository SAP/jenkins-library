package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGaugeExecuteTestsCommand(t *testing.T) {
	t.Parallel()

	testCmd := GaugeExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gaugeExecuteTests", testCmd.Use, "command name incorrect")

}
