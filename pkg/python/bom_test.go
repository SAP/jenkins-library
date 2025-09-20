//go:build unit
// +build unit

package python

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateBOM(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}
	mockFiles := mock.FilesMock{}

	// test
	err := CreateBOM(mockRunner.RunExecutable, mockFiles.FileExists, ".venv", "requirements.txt", "1.2.3", "16")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 2)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"cyclonedx-bom==1.2.3"}, mockRunner.Calls[0].Params)
	assert.Equal(t, ".venv/bin/cyclonedx-py", mockRunner.Calls[1].Exec)
	assert.Equal(t, []string{
		"env",
		"--output-file", "bom-pip.xml",
		"--output-format", "XML",
		"--spec-version", "16"}, mockRunner.Calls[1].Params)
}
