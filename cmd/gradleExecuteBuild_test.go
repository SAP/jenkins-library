package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type gradleExecuteBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (f gradleExecuteBuildMockUtils) DirExists(path string) (bool, error) {
	return strings.EqualFold(path, "path/to/"), nil
}

func (f gradleExecuteBuildMockUtils) FileExists(filePath string) (bool, error) {
	return strings.EqualFold(filePath, "path/to/build.gradle"), nil
}

func (f gradleExecuteBuildMockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return fmt.Errorf("not implemented")
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
			Path: "path/to/project/",
		}
		u := newGradleExecuteBuildTestsUtils()

		err := runGradleExecuteBuild(options, nil, u)
		assert.EqualError(t, err, "the specified gradle build script could not be found")
	})

	t.Run("success case - build.gradle is present", func(t *testing.T) {
		o := &gradleExecuteBuildOptions{
			Path: "path/to/",
		}

		u := newGradleExecuteBuildTestsUtils()

		err := runGradleExecuteBuild(o, nil, u)
		assert.NoError(t, err)
	})

}
