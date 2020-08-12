package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentReadAddonDescriptorCommand(t *testing.T) {

	testCmd := AbapEnvironmentReadAddonDescriptorCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "abapEnvironmentReadAddonDescriptor", testCmd.Use, "command name incorrect")

}
