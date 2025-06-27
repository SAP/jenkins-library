//go:build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAwsS3UploadCommand(t *testing.T) {
	t.Parallel()

	testCmd := AwsS3UploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "awsS3Upload", testCmd.Use, "command name incorrect")

}
