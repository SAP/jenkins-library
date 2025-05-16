//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAscAppUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := AscAppUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "ascAppUpload", testCmd.Use, "command name incorrect")

}
