package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeleteServiceCommand(t *testing.T) {

	testCmd := CloudFoundryDeleteServiceCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "cloudFoundryDeleteService", testCmd.Use, "command name incorrect")

}
