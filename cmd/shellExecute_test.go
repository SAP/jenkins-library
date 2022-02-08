package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type shellExecuteMockUtils struct {
	t      *testing.T
	config *shellExecuteOptions
	*mock.ExecMockRunner
	*mock.FilesMock
}

type shellExecuteFileMock struct {
	*mock.FilesMock
	fileReadContent map[string]string
	fileReadErr     map[string]error
}

func (f *shellExecuteFileMock) FileRead(path string) ([]byte, error) {
	if f.fileReadErr[path] != nil {
		return []byte{}, f.fileReadErr[path]
	}
	return []byte(f.fileReadContent[path]), nil
}

func (f *shellExecuteFileMock) FileExists(path string) (bool, error) {
	return strings.EqualFold(path, "path/to/script/script.sh"), nil
}

func newShellExecuteTestsUtils() shellExecuteMockUtils {
	utils := shellExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func (v *shellExecuteMockUtils) GetConfig() *shellExecuteOptions {
	return v.config
}

func TestRunShellExecute(t *testing.T) {

	t.Run("negative case - script isn't present", func(t *testing.T) {
		c := &shellExecuteOptions{
			Sources: []string{"path/to/script.sh"},
		}
		u := newShellExecuteTestsUtils()

		err := runShellExecute(c, nil, u)
		assert.EqualError(t, err, "the script 'path/to/script.sh' could not be found")
	})

	t.Run("success case - script is present", func(t *testing.T) {
		o := &shellExecuteOptions{}
		u := newShellExecuteTestsUtils()

		err := runShellExecute(o, nil, u)
		assert.NoError(t, err)
	})

	t.Run("success case - script run successfully", func(t *testing.T) {
		o := &shellExecuteOptions{}
		u := newShellExecuteTestsUtils()

		err := runShellExecute(o, nil, u)
		assert.NoError(t, err)
	})

}
