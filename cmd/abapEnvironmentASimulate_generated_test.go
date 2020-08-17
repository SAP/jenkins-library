package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentASimulateCommand(t *testing.T) {

	testCmd := AbapEnvironmentASimulateCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "abapEnvironmentASimulate", testCmd.Use, "command name incorrect")

}
