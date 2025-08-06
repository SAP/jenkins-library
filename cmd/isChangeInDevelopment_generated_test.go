package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsChangeInDevelopmentCommand(t *testing.T) {
	t.Parallel()

	testCmd := IsChangeInDevelopmentCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "isChangeInDevelopment", testCmd.Use, "command name incorrect")

}
