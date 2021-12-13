package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type gradleExecuteBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

type gradleExecuteBuildFileMock struct {
	*mock.FilesMock
	fileReadContent map[string]string
	fileReadErr     map[string]error
}

func (f *gradleExecuteBuildFileMock) FileExists(path string) (bool, error) {
	return strings.EqualFold(path, "path/to/gradle.build"), nil
}

func newGradleExecuteBuildTestsUtils() gradleExecuteBuildMockUtils {
	utils := gradleExecuteBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunGradleExecuteBuild(t *testing.T) {

	t.Run("negative case - build.gradle isn't present", func(t *testing.T) {
		options := &gradleExecuteBuildOptions{
			Path: "path/to/project/build.gradle",
		}
		u := newShellExecuteTestsUtils()

		m := &gradleExecuteBuildFileMock{}

		err := runGradleExecuteBuild(options, nil, u, m)
		assert.EqualError(t, err, "the specified gradle script could not be found")
	})

	t.Run("success case - build.gradle is present", func(t *testing.T) {
		o := &gradleExecuteBuildOptions{
			Path: "path/to/gradle.build",
		}

		u := newGradleExecuteBuildTestsUtils()
		m := &gradleExecuteBuildFileMock{}

		err := runGradleExecuteBuild(o, nil, u, m)
		assert.NoError(t, err)
	})

}
