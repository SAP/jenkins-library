//go:build unit

package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestPublishWithVirtualEnv(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := PublishPackage(mockRunner.RunExecutable, ".venv", "repository", "anything", "anything")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"twine"}, mockRunner.Calls[0].Params)
	assert.Equal(t, ".venv/bin/twine", mockRunner.Calls[1].Exec)
	assert.Equal(t, []string{
		"upload",
		"--username", "anything",
		"--password", "anything",
		"--repository-url", "repository",
		"--disable-progress-bar",
		"dist/*"}, mockRunner.Calls[1].Params)
}
