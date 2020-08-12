package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAAKaaSCheckCVCommand(t *testing.T) {

	testCmd := AAKaaSCheckCVCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "aAKaaSCheckCV", testCmd.Use, "command name incorrect")

}
