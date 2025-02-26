package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type buildahExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newBuildahExecuteTestsUtils() buildahExecuteMockUtils {
	utils := buildahExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunBuildahExecute(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := buildahExecuteOptions{}

		utils := newBuildahExecuteTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runBuildahExecute(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := buildahExecuteOptions{}

		utils := newBuildahExecuteTestsUtils()

		// test
		err := runBuildahExecute(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
