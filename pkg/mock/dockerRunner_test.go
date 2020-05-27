package mock

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDockerExecRunnerRunExecutable(t *testing.T) {
	dockerRunner := DockerExecRunner{
		DockerWorkspace:   "some/path",
		DockerImage:       "image",
		ExecutablesToWrap: []string{"useful-tool"},
	}
	t.Run("tool execution is wrapped", func(t *testing.T) {
		mockRunner := ExecMockRunner{}
		dockerRunner.runner = &mockRunner
		currentDir, err := os.Getwd()
		assert.NoError(t, err)
		err = dockerRunner.RunExecutable("useful-tool", "param", "--flag")
		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockRunner.Calls)) {
			assert.Equal(t, ExecCall{
				Exec:   "docker",
				Params: []string{"run", "--entrypoint=useful-tool", "-v", fmt.Sprintf("%s:some/path", currentDir), "image", "param", "--flag"},
			}, mockRunner.Calls[0])
		}
	})
	t.Run("tool execution is not wrapped", func(t *testing.T) {
		mockRunner := ExecMockRunner{}
		dockerRunner.runner = &mockRunner
		err := dockerRunner.RunExecutable("another-tool", "param", "--flag")
		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockRunner.Calls)) {
			assert.Equal(t, ExecCall{
				Exec:   "another-tool",
				Params: []string{"param", "--flag"},
			}, mockRunner.Calls[0])
		}
	})
	t.Run("error case", func(t *testing.T) {
		mockRunner := ExecMockRunner{}
		mockRunner.ShouldFailOnCommand = map[string]error{}
		mockRunner.ShouldFailOnCommand["some-tool"] = errors.New("failed")
		dockerRunner.runner = &mockRunner
		err := dockerRunner.RunExecutable("some-tool")
		assert.Error(t, err, "failed")
	})
}
