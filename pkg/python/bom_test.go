//go:build unit
// +build unit

package python

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateBOM(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}
	mockFiles := mock.FilesMock{}

	// test
	err := CreateBOM(mockRunner.RunExecutable, mockFiles.FileExists, mockFiles.ReadFile, ".venv", "requirements.txt", "1.2.3", "16")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 3)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"."}, mockRunner.Calls[0].Params)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[1].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"cyclonedx-bom==1.2.3"}, mockRunner.Calls[1].Params)
	assert.Equal(t, ".venv/bin/cyclonedx-py", mockRunner.Calls[2].Exec)
	assert.Equal(t, []string{
		"env",
		"--output-file", "bom-pip.xml",
		"--output-format", "XML",
		"--spec-version", "16"}, mockRunner.Calls[2].Params)
}

func TestCreateBOMWithPyProjectToml(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}
	mockFiles := mock.FilesMock{}

	// Add pyproject.toml with [project] metadata section to mock
	mockFiles.AddFile("pyproject.toml", []byte(`[build-system]
requires = ["setuptools"]

[project]
name = "example-pkg"
version = "0.0.1"
`))

	// test
	err := CreateBOM(mockRunner.RunExecutable, mockFiles.FileExists, mockFiles.ReadFile, ".venv", "requirements.txt", "1.2.3", "1.4")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 3)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[0].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"."}, mockRunner.Calls[0].Params)
	assert.Equal(t, ".venv/bin/pip", mockRunner.Calls[1].Exec)
	assert.Equal(t, []string{
		"install",
		"--upgrade",
		"--root-user-action=ignore",
		"cyclonedx-bom==1.2.3"}, mockRunner.Calls[1].Params)
	assert.Equal(t, ".venv/bin/cyclonedx-py", mockRunner.Calls[2].Exec)
	assert.Equal(t, []string{
		"env",
		"--output-file", "bom-pip.xml",
		"--output-format", "XML",
		"--spec-version", "1.4",
		"--pyproject", "pyproject.toml"}, mockRunner.Calls[2].Params)
}

func TestCreateBOMWithMinimalPyProjectToml(t *testing.T) {
	// init
	mockRunner := mock.ExecMockRunner{}
	mockFiles := mock.FilesMock{}

	// Add pyproject.toml WITHOUT [project] metadata section to mock
	mockFiles.AddFile("pyproject.toml", []byte(`[build-system]
requires = ["setuptools"]
build-backend = "setuptools.build_meta"
`))

	// test
	err := CreateBOM(mockRunner.RunExecutable, mockFiles.FileExists, mockFiles.ReadFile, ".venv", "requirements.txt", "1.2.3", "1.4")

	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 3)
	assert.Equal(t, ".venv/bin/cyclonedx-py", mockRunner.Calls[2].Exec)
	// Should NOT include --pyproject flag since there's no [project] metadata
	assert.Equal(t, []string{
		"env",
		"--output-file", "bom-pip.xml",
		"--output-format", "XML",
		"--spec-version", "1.4"}, mockRunner.Calls[2].Params)
}

func TestPyprojectHasMetadata(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		mockFiles := mock.FilesMock{}
		hasMetadata := pyprojectHasMetadata(mockFiles.ReadFile, "nonexistent.toml")
		assert.False(t, hasMetadata)
	})

	t.Run("file has [project] section", func(t *testing.T) {
		mockFiles := mock.FilesMock{}
		mockFiles.AddFile("pyproject.toml", []byte(`[build-system]
requires = ["setuptools"]

[project]
name = "test"
version = "1.0.0"
`))

		hasMetadata := pyprojectHasMetadata(mockFiles.ReadFile, "pyproject.toml")
		assert.True(t, hasMetadata)
	})

	t.Run("file without [project] section", func(t *testing.T) {
		mockFiles := mock.FilesMock{}
		mockFiles.AddFile("pyproject.toml", []byte(`[build-system]
requires = ["setuptools"]
build-backend = "setuptools.build_meta"
`))

		hasMetadata := pyprojectHasMetadata(mockFiles.ReadFile, "pyproject.toml")
		assert.False(t, hasMetadata)
	})

	t.Run("file with [project] section with whitespace", func(t *testing.T) {
		mockFiles := mock.FilesMock{}
		mockFiles.AddFile("pyproject.toml", []byte(`[build-system]
requires = ["setuptools"]

  [project]
name = "test"
`))

		hasMetadata := pyprojectHasMetadata(mockFiles.ReadFile, "pyproject.toml")
		assert.True(t, hasMetadata)
	})
}
