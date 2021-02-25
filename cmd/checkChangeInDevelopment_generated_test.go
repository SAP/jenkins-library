package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckChangeInDevelopmentCommand(t *testing.T) {
	t.Parallel()

	testCmd := CheckChangeInDevelopmentCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "checkChangeInDevelopment", testCmd.Use, "command name incorrect")

}
