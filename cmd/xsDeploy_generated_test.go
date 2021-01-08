package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXsDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := XsDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "xsDeploy", testCmd.Use, "command name incorrect")

}
