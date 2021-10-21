package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type shellExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newShellExecuteTestsUtils() shellExecuteMockUtils {
	utils := shellExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunShellExecute(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := shellExecuteOptions{}

		utils := newShellExecuteTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runShellExecute(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := shellExecuteOptions{}

		utils := newShellExecuteTestsUtils()

		// test
		err := runShellExecute(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
