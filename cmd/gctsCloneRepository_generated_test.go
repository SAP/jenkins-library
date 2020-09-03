package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsCloneRepositoryCommand(t *testing.T) {

	testCmd := GctsCloneRepositoryCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "gctsCloneRepository", testCmd.Use, "command name incorrect")

}
