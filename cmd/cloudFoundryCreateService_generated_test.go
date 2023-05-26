//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateServiceCommand(t *testing.T) {
	t.Parallel()

	testCmd := CloudFoundryCreateServiceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "cloudFoundryCreateService", testCmd.Use, "command name incorrect")

}
