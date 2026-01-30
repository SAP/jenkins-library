//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsOnboardingStepCommand(t *testing.T) {
	t.Parallel()

	testCmd := FsOnboardingStepCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "fsOnboardingStep", testCmd.Use, "command name incorrect")

}
