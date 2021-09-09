package cnbutils

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuildpackDownload(t *testing.T) {
	t.Run("successfully downloads a buildpack", func(t *testing.T) {
		fileMock := &CnbFileMockUtils{
			FilesMock: &mock.FilesMock{},
		}
		fileMock.AddDir("/tmp/test-dir")
		_, err := DownloadBuildpacks("/test", []string{"test"}, &DockerMock{}, fileMock)

		assert.NoError(t, err)
		assert.True(t, fileMock.HasRemovedFile("/tmp/test-dir"))
	})
}

func TestBuildpackCopy(t *testing.T) {
	t.Run("successfully downloads a buildpack", func(t *testing.T) {

		err := copyBuildPack("/src", "/dst", &CnbFileMockUtils{
			FilesMock: &mock.FilesMock{},
		})

		assert.NoError(t, err)
	})
}
