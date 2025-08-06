package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGcpPublishEventCommand(t *testing.T) {
	t.Parallel()

	testCmd := GcpPublishEventCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gcpPublishEvent", testCmd.Use, "command name incorrect")

}
