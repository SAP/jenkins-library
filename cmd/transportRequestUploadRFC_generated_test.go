//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestUploadRFCCommand(t *testing.T) {
	t.Parallel()

	testCmd := TransportRequestUploadRFCCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestUploadRFC", testCmd.Use, "command name incorrect")

}
