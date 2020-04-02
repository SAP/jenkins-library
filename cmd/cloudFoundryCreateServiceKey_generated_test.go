package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateServiceKeyCommand(t *testing.T) {

	testCmd := CloudFoundryCreateServiceKeyCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "cloudFoundryCreateServiceKey", testCmd.Use, "command name incorrect")

}
