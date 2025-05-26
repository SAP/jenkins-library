//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiProxyListCommand(t *testing.T) {
	t.Parallel()

	testCmd := ApiProxyListCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "apiProxyList", testCmd.Use, "command name incorrect")

}
