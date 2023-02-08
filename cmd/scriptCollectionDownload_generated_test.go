package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScriptCollectionDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := ScriptCollectionDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "scriptCollectionDownload", testCmd.Use, "command name incorrect")

}
