package cnbutils_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	fakeImage "github.com/google/go-containerregistry/pkg/v1/fake"
	"github.com/stretchr/testify/assert"
)

func TestBuildpackDownload(t *testing.T) {
	var mockUtils = &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		DownloadMock:   &mock.DownloadMock{},
	}

	t.Run("it creates an order object", func(t *testing.T) {
		fakeImg := &fakeImage.FakeImage{}
		fakeImg.ConfigFileReturns(&v1.ConfigFile{
			Config: v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"testbuildpack\", \"version\": \"0.0.1\"}",
				},
			},
		}, nil)
		mockUtils.ReturnImage = fakeImg
		mockUtils.RemoteImageInfo = fakeImg

		order, err := cnbutils.DownloadBuildpacks("/destination", []string{"buildpack"}, "/tmp/config.json", mockUtils)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(order.Order))
	})
}
