package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsCloneRepositoryCommand(t *testing.T) {

	testCmd := GctsCloneRepositoryCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gctsCloneRepository", testCmd.Use, "command name incorrect")

}
