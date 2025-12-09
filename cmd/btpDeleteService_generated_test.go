//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtpDeleteServiceCommand(t *testing.T) {
	t.Parallel()

	testCmd := BtpDeleteServiceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "btpDeleteService", testCmd.Use, "command name incorrect")

}
