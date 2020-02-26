package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckSpotBugsCommand(t *testing.T) {

	testCmd := CheckSpotBugsCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "checkSpotBugs", testCmd.Use, "command name incorrect")

}
