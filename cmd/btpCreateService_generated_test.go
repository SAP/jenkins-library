//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtpCreateServiceCommand(t *testing.T) {
	t.Parallel()

	testCmd := BtpCreateServiceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "btpCreateService", testCmd.Use, "command name incorrect")

}
