package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type pythonBuildMockUtils struct {
	t      *testing.T
	config *pythonBuildOptions
	*mock.ExecMockRunner
	*mock.FilesMock
}

type pythonBuildFileMock struct {
	*mock.FilesMock
	dirPathContent map[string]string
	dirPathErr     map[string]error
}

func newPythonBuildTestsUtils() pythonBuildMockUtils {
	utils := pythonBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func (f *pythonBuildFileMock) FileExists(path string) (bool, error) {
	return strings.EqualFold(path, "path/to/dir/python-project"), nil
}

func (f *pythonBuildFileMock) DirExists(path string) (bool, error) {
	return strings.EqualFold(path, "path/to/dir/python-project"), nil
}

func (f *pythonBuildMockUtils) GetConfig() *pythonBuildOptions {
	return f.config
}

func TestRunPythonBuild(t *testing.T) {
	t.Run("negative case - python project is not present", func(t *testing.T) {
		c := &pythonBuildOptions{
			Sources: []string{"path/to/python-project"},
		}
		u := newPythonBuildTestsUtils()
		err := runPythonBuild(c, nil, u)
		assert.EqualError(t, err, "the python project dir 'path/to/python-project' could not be found")
	})

	t.Run("success case - python project is present", func(t *testing.T) {
		o := &pythonBuildOptions{}
		u := newPythonBuildTestsUtils()

		err := runPythonBuild(o, nil, u)
		assert.NoError(t, err)
	})

	t.Run("success case - python project build successfully", func(t *testing.T) {
		o := &pythonBuildOptions{}
		u := newPythonBuildTestsUtils()

		err := runPythonBuild(o, nil, u)
		assert.NoError(t, err)
	})

}
