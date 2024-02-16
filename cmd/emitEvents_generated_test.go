//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmitEventsCommand(t *testing.T) {
	t.Parallel()

	testCmd := EmitEventsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "emitEvents", testCmd.Use, "command name incorrect")

}
