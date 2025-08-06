package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGradleExecuteBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := GradleExecuteBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "gradleExecuteBuild", testCmd.Use, "command name incorrect")
}
