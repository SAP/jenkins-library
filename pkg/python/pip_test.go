package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

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
