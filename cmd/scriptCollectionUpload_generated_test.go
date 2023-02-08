package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScriptCollectionUploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ScriptCollectionUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "scriptCollectionUpload", testCmd.Use, "command name incorrect")

}
