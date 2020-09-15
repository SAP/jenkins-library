package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryFaasDeployCommand(t *testing.T) {

	testCmd := CloudFoundryFaasDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "cloudFoundryFaasDeploy", testCmd.Use, "command name incorrect")

}
