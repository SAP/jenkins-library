//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := CloudFoundryDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "cloudFoundryDeploy", testCmd.Use, "command name incorrect")

}
