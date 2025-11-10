//go:build unit
// +build unit

package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateVirtualEnvironment(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}
	mockFiles := mock.FilesMock{}

	// test
	_, err := CreateVirtualEnvironment(mockRunner.RunExecutable, mockFiles.RemoveAll, ".venv")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 3)
	assert.Equal(t, "python3", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{"-m", "venv", ".venv"}, mockRunner.Calls[0].Params)
	assert.Equal(t, "bash", mockRunner.Calls[1].Exec)
	assert.Equal(t, []string{"-c", "source .venv/bin/activate"}, mockRunner.Calls[1].Params)
}
