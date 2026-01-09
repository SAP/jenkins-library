//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtpDeleteServiceInstanceCommand(t *testing.T) {
	t.Parallel()

	testCmd := BtpDeleteServiceInstanceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "btpDeleteServiceInstance", testCmd.Use, "command name incorrect")

}
