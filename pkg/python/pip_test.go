package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestInstallProjectDependencies(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallProjectDependencies(mockRunner.RunExecutable, "pip")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "."}, mockRunner.Calls[0].Params)
}

func TestInstallBuild(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallBuild(mockRunner.RunExecutable, "pip")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "build"}, mockRunner.Calls[0].Params)
}

func TestInstallWheel(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallWheel(mockRunner.RunExecutable, "pip")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, mockRunner.Calls[0].Params)
}

func TestInstallPip(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallPip(mockRunner.RunExecutable, "pip")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "pip"}, mockRunner.Calls[0].Params)
}
