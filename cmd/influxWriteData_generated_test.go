package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfluxWriteDataCommand(t *testing.T) {
	t.Parallel()

	testCmd := InfluxWriteDataCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "influxWriteData", testCmd.Use, "command name incorrect")

}
