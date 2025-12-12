//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMtaBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := MtaBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "mtaBuild", testCmd.Use, "command name incorrect")

}
