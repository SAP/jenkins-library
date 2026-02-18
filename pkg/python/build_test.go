//go:build unit

package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuildWithSetupPy(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}
	buildFlags := []string{"--verbose"}
	setupFlags := []string{"egg_info", "--tag-build=pr13"}

	// test
	err := BuildWithSetupPy(mockRunner.RunExecutable, ".venv", buildFlags, setupFlags)

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 2)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"wheel"}, mockRunner.Calls[0].Params)
	assert.Equal(t, ".venv/bin/python", mockRunner.Calls[1].Exec)
	assert.Equal(t, []string{
		"--verbose",
		"setup.py",
		"egg_info",
		"--tag-build=pr13",
		"sdist",
		"bdist_wheel",
	}, mockRunner.Calls[1].Params)
}

func TestBuildWithPyProjectToml(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := BuildWithPyProjectToml(mockRunner.RunExecutable, "", nil, nil)

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 5)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"pip"}, mockRunner.Calls[0].Params)
	assert.Equal(t, "pip", mockRunner.Calls[1].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"."}, mockRunner.Calls[1].Params)
	assert.Equal(t, "pip", mockRunner.Calls[2].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"build"}, mockRunner.Calls[2].Params)
	assert.Equal(t, "pip", mockRunner.Calls[3].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"wheel"}, mockRunner.Calls[3].Params)
	assert.Equal(t, "python", mockRunner.Calls[4].Exec)
	assert.Equal(t, []string{
		"-m", "build",
		"--no-isolation"}, mockRunner.Calls[4].Params)
}

func TestBuildWithPyProjectTomlWithVirtualEnv(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := BuildWithPyProjectToml(mockRunner.RunExecutable, ".venv", nil, nil)

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 5)
	assert.Equal(t, ".venv/bin/python", mockRunner.Calls[4].Exec)
	assert.Equal(t, []string{
		"-m", "build",
		"--no-isolation"}, mockRunner.Calls[4].Params)
}
