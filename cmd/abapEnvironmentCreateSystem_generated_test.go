package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentCreateSystemCommand(t *testing.T) {

	testCmd := AbapEnvironmentCreateSystemCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentCreateSystem", testCmd.Use, "command name incorrect")

}
