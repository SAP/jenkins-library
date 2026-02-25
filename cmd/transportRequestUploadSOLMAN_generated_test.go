//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestUploadSOLMANCommand(t *testing.T) {
	t.Parallel()

	testCmd := TransportRequestUploadSOLMANCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestUploadSOLMAN", testCmd.Use, "command name incorrect")

}
