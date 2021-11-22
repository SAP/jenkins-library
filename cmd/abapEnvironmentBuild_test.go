package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type abapEnvironmentBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAbapEnvironmentBuildTestsUtils() abapEnvironmentBuildMockUtils {
	utils := abapEnvironmentBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunAbapEnvironmentBuild(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		/*
			// init
			config := abapEnvironmentBuildOptions{}

			utils := newAbapEnvironmentBuildTestsUtils()
			utils.AddFile("file.txt", []byte("dummy content"))

			// test
			err := runAbapEnvironmentBuild(&config, nil, utils)

			// assert
			assert.NoError(t, err)
		*/
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		/*
			// init
			config := abapEnvironmentBuildOptions{}

			utils := newAbapEnvironmentBuildTestsUtils()


			// test
			err := runAbapEnvironmentBuild(&config, nil, utils)

			// assert
			assert.EqualError(t, err, "cannot run without important file")
		*/
	})
}
