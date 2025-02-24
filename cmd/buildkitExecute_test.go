package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type buildkitExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newBuildkitExecuteTestsUtils() buildkitExecuteMockUtils {
	utils := buildkitExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunBuildkitExecute(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := buildkitExecuteOptions{}

		utils := newBuildkitExecuteTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runBuildkitExecute(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := buildkitExecuteOptions{}

		utils := newBuildkitExecuteTestsUtils()

		// test
		err := runBuildkitExecute(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
