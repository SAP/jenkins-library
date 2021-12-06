package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type gradleExecuteBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newGradleExecuteBuildTestsUtils() gradleExecuteBuildMockUtils {
	utils := gradleExecuteBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunGradleExecuteBuild(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := gradleExecuteBuildOptions{}

		utils := newGradleExecuteBuildTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runGradleExecuteBuild(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := gradleExecuteBuildOptions{}

		utils := newGradleExecuteBuildTestsUtils()

		// test
		err := runGradleExecuteBuild(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
