package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUiVeri5ExecuteTestsCommand(t *testing.T) {

	testCmd := UiVeri5ExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "uiVeri5ExecuteTests", testCmd.Use, "command name incorrect")

}
