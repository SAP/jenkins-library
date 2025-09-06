package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}

	// test
	err := Build(mockRunner.RunExecutable, "python", nil, nil)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "python", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{"-m", "build", "--no-isolation"}, mockRunner.Calls[0].Params)
}
