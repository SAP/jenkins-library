//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsCreateRepositoryCommand(t *testing.T) {
	t.Parallel()

	testCmd := GctsCreateRepositoryCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gctsCreateRepository", testCmd.Use, "command name incorrect")

}
