//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type helloExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (h *helloExecuteMockUtils) GetDockerImageValue(stepName string) (string, error) {
	return "alpine:3.18", nil
}

func newHelloExecuteTestsUtils() *helloExecuteMockUtils {
	utils := &helloExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunHelloExecute(t *testing.T) {
	t.Parallel()

	t.Run("happy path with default username", func(t *testing.T) {
		t.Parallel()
		// init
		config := helloExecuteOptions{HelloUsername: "World"}

		utils := newHelloExecuteTestsUtils()

		// test
		err := runHelloExecute(&config, nil, utils)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, "docker", utils.Calls[0].Exec)
		assert.Equal(t, []string{"run", "--rm", "alpine:3.18", "sh", "-c", "echo 'Hello World!'"}, utils.Calls[0].Params)
	})

	t.Run("happy path with custom username", func(t *testing.T) {
		t.Parallel()
		// init
		config := helloExecuteOptions{HelloUsername: "Alice"}

		utils := newHelloExecuteTestsUtils()

		// test
		err := runHelloExecute(&config, nil, utils)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, "docker", utils.Calls[0].Exec)
		assert.Equal(t, []string{"run", "--rm", "alpine:3.18", "sh", "-c", "echo 'Hello Alice!'"}, utils.Calls[0].Params)
	})

	t.Run("error when docker command fails", func(t *testing.T) {
		t.Parallel()
		// init
		config := helloExecuteOptions{HelloUsername: "Bob"}

		utils := newHelloExecuteTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{
			"docker": assert.AnError,
		}

		// test
		err := runHelloExecute(&config, nil, utils)

		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute greeting command in docker")
	})
}
