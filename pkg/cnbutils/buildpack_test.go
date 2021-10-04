package cnbutils

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

var mockUtils = MockUtils{
	ExecMockRunner: &mock.ExecMockRunner{},
	FilesMock:      &mock.FilesMock{},
	DockerMock:     &DockerMock{},
}

func TestBuildpackDownload(t *testing.T) {
	t.Run("successfully downloads a buildpack", func(t *testing.T) {
		mockUtils.AddDir("/tmp/testtest")
		_, err := DownloadBuildpacks("/test", []string{"test"}, "/test/config.json", mockUtils)

		assert.NoError(t, err)
		assert.True(t, mockUtils.HasRemovedFile("/tmp/testtest"))
	})
}

func TestBuildpackCopy(t *testing.T) {
	t.Run("successfully downloads a buildpack", func(t *testing.T) {

		mockUtils.AddDir("/src/buildpack/0.0.1")
		mockUtils.AddDir("/dst")
		err := copyBuildPack("/src", "/dst", mockUtils)

		assert.NoError(t, err)
	})
}
