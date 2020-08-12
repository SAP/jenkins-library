package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAAKaaSCheckPVCommand(t *testing.T) {

	testCmd := AAKaaSCheckPVCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "aAKaaSCheckPV", testCmd.Use, "command name incorrect")

}
