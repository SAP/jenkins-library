package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaPatchCommand(t *testing.T) {

	testCmd := SchemaPatchCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "schemaPatch", testCmd.Use, "command name incorrect")

}
