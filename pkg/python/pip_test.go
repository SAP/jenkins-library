//go:build unit
// +build unit

package python

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestInstallPip(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallPip(mockRunner.RunExecutable, "")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "pip"}, mockRunner.Calls[0].Params)
}

func TestInstallProjectDependencies(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallProjectDependencies(mockRunner.RunExecutable, "")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"."}, mockRunner.Calls[0].Params)
}

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

func TestInstallBuild(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallBuild(mockRunner.RunExecutable, "")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"build"}, mockRunner.Calls[0].Params)
}

func TestInstallBuildWithVirtualEnv(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallBuild(mockRunner.RunExecutable, ".venv")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"build"}, mockRunner.Calls[0].Params)
}

func TestInstallWheel(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := InstallWheel(mockRunner.RunExecutable, "")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 1)
	assert.Equal(t, "pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"wheel"}, mockRunner.Calls[0].Params)
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

func TestInstallTestDependencies(t *testing.T) {
	tests := []struct {
		name       string
		virtualEnv string
		wantExec   string
	}{
		{name: "no virtualenv", virtualEnv: "", wantExec: "pip"},
		{name: "with virtualenv", virtualEnv: ".venv", wantExec: ".venv/bin/pip"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := mock.ExecMockRunner{}

			err := InstallTestDependencies(mockRunner.RunExecutable, tt.virtualEnv)

			assert.NoError(t, err)
			assert.Len(t, mockRunner.Calls, 2)
			assert.Equal(t, tt.wantExec, mockRunner.Calls[0].Exec)
			assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "pytest"}, mockRunner.Calls[0].Params)
			assert.Equal(t, tt.wantExec, mockRunner.Calls[1].Exec)
			assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "pytest-cov"}, mockRunner.Calls[1].Params)
		})
	}
}

func TestInstallTestDependenciesPytestFailure(t *testing.T) {
	t.Parallel()
	mockRunner := mock.ExecMockRunner{
		ShouldFailOnCommand: map[string]error{"pytest$": fmt.Errorf("pip install failed")},
	}

	err := InstallTestDependencies(mockRunner.RunExecutable, "")

	assert.Error(t, err)
	assert.Len(t, mockRunner.Calls, 1, "pytest-cov install must not be attempted after pytest install fails")
}
