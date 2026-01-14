//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtpCreateServiceBindingCommand(t *testing.T) {
	t.Parallel()

	testCmd := BtpCreateServiceBindingCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "btpCreateServiceBinding", testCmd.Use, "command name incorrect")

}
