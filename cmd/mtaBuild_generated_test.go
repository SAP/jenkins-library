package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMtaBuildCommand(t *testing.T) {

	testCmd := MtaBuildCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "mtaBuild", testCmd.Use, "command name incorrect")

}
