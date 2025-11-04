//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBOMCommand(t *testing.T) {
	t.Parallel()

	testCmd := ValidateBOMCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "validateBOM", testCmd.Use, "command name incorrect")

}
