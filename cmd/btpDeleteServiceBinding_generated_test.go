//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtpDeleteServiceBindingCommand(t *testing.T) {
	t.Parallel()

	testCmd := BtpDeleteServiceBindingCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "btpDeleteServiceBinding", testCmd.Use, "command name incorrect")

}
