//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbapEnvironmentPushATCSystemConfigCommand(t *testing.T) {
	t.Parallel()

	testCmd := AbapEnvironmentPushATCSystemConfigCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "abapEnvironmentPushATCSystemConfig", testCmd.Use, "command name incorrect")

}
