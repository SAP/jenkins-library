package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitopsUpdateDeploymentCommand(t *testing.T) {

	testCmd := GitopsUpdateDeploymentCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gitopsUpdateDeployment", testCmd.Use, "command name incorrect")

}
