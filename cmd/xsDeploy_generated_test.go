package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXsDeployCommand(t *testing.T) {

	testCmd := XsDeployCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "xsDeploy", testCmd.Use, "command name incorrect")

}
