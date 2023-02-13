package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPackageListCommand(t *testing.T) {
	t.Parallel()

	testCmd := GetPackageListCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "getPackageList", testCmd.Use, "command name incorrect")

}
