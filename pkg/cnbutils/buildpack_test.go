package cnbutils_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuildpackDownload(t *testing.T) {
	var mockUtils = &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}

	t.Run("successfully downloads a buildpack", func(t *testing.T) {
		mockUtils.AddDir("/tmp/testtest")
		_, err := cnbutils.DownloadBuildpacks("/test", []string{"test"}, "/test/config.json", mockUtils)

		assert.NoError(t, err)
		assert.True(t, mockUtils.HasRemovedFile("/tmp/testtest"))
	})
}
