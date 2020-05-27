package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsDeployCommand(t *testing.T) {

	testCmd := GctsDeployCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "gctsDeploy", testCmd.Use, "command name incorrect")

}
