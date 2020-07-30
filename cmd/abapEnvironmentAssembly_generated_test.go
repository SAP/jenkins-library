package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentAssemblyCommand(t *testing.T) {

	testCmd := AbapEnvironmentAssemblyCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "abapEnvironmentAssembly", testCmd.Use, "command name incorrect")

}
