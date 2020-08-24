package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateServiceCommand(t *testing.T) {

	testCmd := CloudFoundryCreateServiceCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "cloudFoundryCreateService", testCmd.Use, "command name incorrect")

}
