package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSonarExecuteScanCommand(t *testing.T) {

	testCmd := SonarExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "sonarExecuteScan", testCmd.Use, "command name incorrect")

}
