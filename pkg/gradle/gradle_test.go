//go:build unit

package gradle

import (
	"errors"
	"os"

	"github.com/SAP/jenkins-library/pkg/mock"

	"testing"

	"github.com/stretchr/testify/assert"
)

type MockUtils struct {
	existingFiles []string
	writtenFiles  []string
	removedFiles  []string
	*mock.FilesMock
	*mock.ExecMockRunner
}

func (m *MockUtils) FileExists(filePath string) (bool, error) {
	for _, filename := range m.existingFiles {
		if filename == filePath {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUtils) FileWrite(path string, content []byte, perm os.FileMode) error {
	m.writtenFiles = append(m.writtenFiles, path)
	return nil
}

func (m *MockUtils) FileRemove(path string) error {
	m.removedFiles = append(m.removedFiles, path)
	return nil
}

func TestExecute(t *testing.T) {
	t.Run("success - run command use gradle tool", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"path/to/build.gradle"},
		}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "build",
			InitScriptContent: "",
			UseWrapper:        false,
			ProjectProperties: map[string]string{"propName": "propValue"},
		}

		_, err := Execute(&opts, utils)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to", "-PpropName=propValue"}}, utils.Calls[0])
		assert.Equal(t, []string(nil), utils.writtenFiles)
		assert.Equal(t, []string(nil), utils.removedFiles)
	})

	t.Run("success - run command use gradlew tool", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"path/to/build.gradle", "gradlew"},
		}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "build",
			InitScriptContent: "",
			UseWrapper:        true,
		}

		_, err := Execute(&opts, utils)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "./gradlew", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, []string(nil), utils.writtenFiles)
		assert.Equal(t, []string(nil), utils.removedFiles)
	})

	t.Run("use init script to apply plugin", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"path/to/build.gradle.kts"},
		}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "build",
			InitScriptContent: "some content",
			UseWrapper:        false,
		}

		_, err := Execute(&opts, utils)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "gradle", Params: []string{"build", "-p", "path/to", "--init-script", "initScript.gradle.tmp"}}, utils.Calls[1])
		assert.Equal(t, []string{"initScript.gradle.tmp"}, utils.writtenFiles)
		assert.Equal(t, []string{"initScript.gradle.tmp"}, utils.removedFiles)
	})

	t.Run("failed - use init script to apply plugin", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock: &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"gradle tasks -p path/to": errors.New("failed to get tasks")},
			},
			existingFiles: []string{"path/to/build.gradle.kts"},
		}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "build",
			InitScriptContent: "some content",
			UseWrapper:        false,
		}

		_, err := Execute(&opts, utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get tasks")
	})

	t.Run("use init script to apply an existing plugin", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"path/to/build.gradle.kts"},
		}
		utils.StdoutReturn = map[string]string{"gradle tasks -p path/to": "createBom"}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "createBom",
			InitScriptContent: "some content",
			UseWrapper:        false,
		}

		_, err := Execute(&opts, utils)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "gradle", Params: []string{"createBom", "-p", "path/to"}}, utils.Calls[1])
		assert.Equal(t, []string(nil), utils.writtenFiles)
		assert.Equal(t, []string(nil), utils.removedFiles)
	})

	t.Run("failed - run command", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock: &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"gradle build -p path/to": errors.New("failed to build")},
			},
			existingFiles: []string{"path/to/build.gradle"},
		}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "build",
			InitScriptContent: "",
			UseWrapper:        false,
		}

		_, err := Execute(&opts, utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to build")
	})

	t.Run("failed - missing gradle build script", func(t *testing.T) {
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{},
		}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "build",
			InitScriptContent: "",
			UseWrapper:        false,
		}

		_, err := Execute(&opts, utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the specified gradle build script could not be found")
	})

	t.Run("success - should return stdOut", func(t *testing.T) {
		expectedOutput := "mocked output"
		utils := &MockUtils{
			FilesMock:      &mock.FilesMock{},
			ExecMockRunner: &mock.ExecMockRunner{},
			existingFiles:  []string{"path/to/build.gradle"},
		}
		utils.StdoutReturn = map[string]string{"gradle build -p path/to": expectedOutput}
		opts := ExecuteOptions{
			BuildGradlePath:   "path/to",
			Task:              "build",
			InitScriptContent: "",
			UseWrapper:        false,
		}

		actualOutput, err := Execute(&opts, utils)
		assert.NoError(t, err)

		assert.Equal(t, expectedOutput, actualOutput)
	})
}
