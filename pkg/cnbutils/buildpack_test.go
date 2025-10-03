package cnbutils_test

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
	fakeImage "github.com/google/go-containerregistry/pkg/v1/fake"
	"github.com/stretchr/testify/assert"
)

func TestBuildpackDownload(t *testing.T) {
	var mockUtils = &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		DownloadMock:   &mock.DownloadMock{},
	}

	t.Run("successfully downloads a buildpack", func(t *testing.T) {
		fakeImg := &fakeImage.FakeImage{}
		fakeImg.DigestReturns(v1.NewHash("sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"))
		fakeImg.ConfigFileReturns(&v1.ConfigFile{
			Config: v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"testbuildpack\", \"version\": \"0.0.1\"}",
				},
			},
		}, nil)
		mockUtils.ReturnImage = fakeImg
		mockUtils.RemoteImageInfo = fakeImg

		err := cnbutils.DownloadBuildpacks("/destination", []string{"buildpack"}, "/tmp/config.json", mockUtils)
		assert.NoError(t, err)
	})
}

func TestGetMetadata(t *testing.T) {
	var mockUtils = &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		DownloadMock: &mock.DownloadMock{
			ImageInfoStub: func(imageRef string) (v1.Image, error) {
				return &fake.FakeImage{
					ConfigFileStub: func() (*v1.ConfigFile, error) {
						return &v1.ConfigFile{
							Config: v1.Config{
								Labels: map[string]string{
									"io.buildpacks.buildpackage.metadata": fmt.Sprintf("{\"id\": \"%s\", \"version\": \"0.0.1\"}", imageRef),
								},
							},
						}, nil
					},
				}, nil
			},
		},
	}

	t.Run("returns empty metadata", func(t *testing.T) {
		meta, err := cnbutils.GetMetadata(nil, mockUtils)
		assert.NoError(t, err)
		assert.Empty(t, meta)
	})

	t.Run("returns metadata of the provided buildpacks", func(t *testing.T) {
		meta, err := cnbutils.GetMetadata([]string{"buildpack1", "buildpack2"}, mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, []cnbutils.BuildPackMetadata{
			{
				ID:      "buildpack1",
				Version: "0.0.1",
			},
			{
				ID:      "buildpack2",
				Version: "0.0.1",
			},
		}, meta)
	})
}
