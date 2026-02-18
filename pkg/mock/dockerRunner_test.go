//go:build unit

package mock

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerExecRunnerAddExecConfig(t *testing.T) {
	t.Run("no executable provided results in error", func(t *testing.T) {
		dockerRunner := DockerExecRunner{}
		err := dockerRunner.AddExecConfig("", DockerExecConfig{})
		assert.Error(t, err, "'executable' needs to be provided")
		assert.Nil(t, dockerRunner.executablesToWrap)
	})
	t.Run("no image provided results in error", func(t *testing.T) {
		dockerRunner := DockerExecRunner{}
		err := dockerRunner.AddExecConfig("useful-tool", DockerExecConfig{})
		assert.Error(t, err, "the DockerExecConfig must specify a docker image")
	})
	t.Run("success case", func(t *testing.T) {
		dockerRunner := DockerExecRunner{}
		config := DockerExecConfig{Image: "image", Workspace: "/var/home"}
		err := dockerRunner.AddExecConfig("useful-tool", config)
		assert.NoError(t, err)
		if assert.NotNil(t, dockerRunner.executablesToWrap) {
			assert.Len(t, dockerRunner.executablesToWrap, 1)
			assert.Equal(t, config, dockerRunner.executablesToWrap["useful-tool"])
		}
	})
}

func TestDockerExecRunnerRunExecutable(t *testing.T) {
	dockerRunner := DockerExecRunner{}
	_ = dockerRunner.AddExecConfig("useful-tool", DockerExecConfig{
		Image:     "image",
		Workspace: "some/path",
	})
	t.Run("tool execution is wrapped", func(t *testing.T) {
		mockRunner := ExecMockRunner{}
		dockerRunner.Runner = &mockRunner
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
		dockerRunner.Runner = &mockRunner
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
		dockerRunner.Runner = &mockRunner
		err := dockerRunner.RunExecutable("some-tool")
		assert.Error(t, err, "failed")
	})
}
