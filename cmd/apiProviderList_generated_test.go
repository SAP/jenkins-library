package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiProviderListCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiProviderListCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiProviderList", testCmd.Use, "command name incorrect")

}
