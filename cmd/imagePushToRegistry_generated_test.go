//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImagePushToRegistryCommand(t *testing.T) {
	t.Parallel()

	testCmd := ImagePushToRegistryCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "imagePushToRegistry", testCmd.Use, "command name incorrect")

}
