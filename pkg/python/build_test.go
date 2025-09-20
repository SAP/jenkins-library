//go:build unit
// +build unit

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
	assert.Len(t, mockRunner.Calls, 1)
	assert.Equal(t, ".venv/bin/python", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"--verbose",
		"setup.py",
		"egg_info",
		"--tag-build=pr13",
		"sdist",
		"bdist_wheel",
	}, mockRunner.Calls[0].Params)
}
