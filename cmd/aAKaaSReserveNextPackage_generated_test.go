package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAAKaaSReserveNextPackageCommand(t *testing.T) {

	testCmd := AAKaaSReserveNextPackageCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "aAKaaSReserveNextPackage", testCmd.Use, "command name incorrect")

}
