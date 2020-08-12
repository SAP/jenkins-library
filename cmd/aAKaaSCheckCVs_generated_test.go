package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAAKaaSCheckCVsCommand(t *testing.T) {

	testCmd := AAKaaSCheckCVsCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "aAKaaSCheckCVs", testCmd.Use, "command name incorrect")

}
