package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type packBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newPackBuildTestsUtils() packBuildMockUtils {
	utils := packBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunPackBuild(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := packBuildOptions{}

		utils := newPackBuildTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runPackBuild(&config, nil, utils, &packBuildCommonPipelineEnvironment{})

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := packBuildOptions{}

		utils := newPackBuildTestsUtils()

		// test
		err := runPackBuild(&config, nil, utils, &packBuildCommonPipelineEnvironment{})

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
