//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiProxyUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiProxyUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiProxyUpload", testCmd.Use, "command name incorrect")

}
