package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPmdCommand(t *testing.T) {

	testCmd := CheckPmdCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "checkPmd", testCmd.Use, "command name incorrect")

}
