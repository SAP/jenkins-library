//go:build unit
// +build unit

package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestInstallRequirements(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallRequirements(mockRunner.RunExecutable, "", "requirements.txt")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"--requirement", "requirements.txt"}, mockRunner.Calls[0].Params)
}

func TestInstallTwine(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallTwine(mockRunner.RunExecutable, "")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 1)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"twine"}, mockRunner.Calls[0].Params)
}

func TestInstallCycloneDXWithVersion(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallCycloneDX(mockRunner.RunExecutable, "", "1.0.0")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 1)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"cyclonedx-bom==1.0.0"}, mockRunner.Calls[0].Params)
}
