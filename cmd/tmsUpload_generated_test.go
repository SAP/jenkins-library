//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTmsUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := TmsUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "tmsUpload", testCmd.Use, "command name incorrect")

}
