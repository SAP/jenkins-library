//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineCreateScanSummaryCommand(t *testing.T) {
	t.Parallel()

	testCmd := PipelineCreateScanSummaryCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "pipelineCreateScanSummary", testCmd.Use, "command name incorrect")

}
