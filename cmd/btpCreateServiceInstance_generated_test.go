//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtpCreateServiceInstanceCommand(t *testing.T) {
	t.Parallel()

	testCmd := BtpCreateServiceInstanceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "btpCreateServiceInstance", testCmd.Use, "command name incorrect")

}
