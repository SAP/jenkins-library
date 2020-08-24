package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeployCommand(t *testing.T) {

	testCmd := CloudFoundryDeployCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "cloudFoundryDeploy", testCmd.Use, "command name incorrect")

}
