package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTestExecuteTestsUtils() testExecuteMockUtils {
	utils := testExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunTestExecute(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// init
		config := testExecuteOptions{}

		utils := newTestExecuteTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runTestExecute(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		// init
		config := testExecuteOptions{}

		utils := newTestExecuteTestsUtils()

		// test
		err := runTestExecute(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
