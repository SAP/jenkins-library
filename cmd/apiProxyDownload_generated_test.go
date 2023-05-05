//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiProxyDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiProxyDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiProxyDownload", testCmd.Use, "command name incorrect")

}
