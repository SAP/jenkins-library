//go:build unit
// +build unit

package cnbutils_test

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
	"github.com/stretchr/testify/assert"
)

func TestOrderSave(t *testing.T) {
	t.Run("successfully Encode struct to toml format (multiple buildpacks)", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}

		testBuildpacks := []cnbutils.BuildPackMetadata{
			{
				ID:      "paketo-buildpacks/sap-machine",
				Version: "1.1.1",
			},
			{
				ID:      "paketo-buildpacks/java",
				Version: "2.2.2",
			},
		}
		testOrder := cnbutils.Order{
			Utils: mockUtils,
		}

		var testEntry cnbutils.OrderEntry
		testEntry.Group = append(testEntry.Group, testBuildpacks...)
		testOrder.Order = []cnbutils.OrderEntry{testEntry}

		err := testOrder.Save("/tmp/order.toml")

		assert.NoError(t, err)
		assert.True(t, mockUtils.HasWrittenFile("/tmp/order.toml"))
		result, err := mockUtils.FileRead("/tmp/order.toml")
		assert.NoError(t, err)
		assert.Equal(t, "[[order]]\n\n  [[order.group]]\n    id = \"paketo-buildpacks/sap-machine\"\n    version = \"1.1.1\"\n\n  [[order.group]]\n    id = \"paketo-buildpacks/java\"\n    version = \"2.2.2\"\n", string(result))
	})

	t.Run("raises an error if unable to write the file", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		mockUtils.FileWriteErrors = map[string]error{
			"/tmp/order.toml": fmt.Errorf("unable to write to file"),
		}
		testOrder := cnbutils.Order{
			Utils: mockUtils,
		}

		err := testOrder.Save("/tmp/order.toml")

		assert.Error(t, err, "unable to write to file")
		assert.False(t, mockUtils.HasWrittenFile("/tmp/order.toml"))
	})
}

func TestCreateOrder(t *testing.T) {
	imageStub := func(imageRef, target string) (v1.Image, error) {
		fakeImage := &fake.FakeImage{}
		var imageConfig v1.Config
		switch imageRef {
		case "pre-buildpack":
			imageConfig = v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"pre-testbuildpack\", \"version\": \"0.0.1\"}",
				},
			}
		case "post-buildpack":
			imageConfig = v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"post-testbuildpack\", \"version\": \"0.0.1\"}",
				},
			}
		default:
			imageConfig = v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"testbuildpack\", \"version\": \"0.0.1\"}",
				},
			}
		}

		fakeImage.ConfigFileReturns(&v1.ConfigFile{
			Config: imageConfig,
		}, nil)

		return fakeImage, nil
	}

	mockUtils := &cnbutils.MockUtils{
		FilesMock: &mock.FilesMock{},
		DownloadMock: &mock.DownloadMock{
			ImageContentStub: imageStub,
			ImageInfoStub: func(imageRef string) (v1.Image, error) {
				return imageStub(imageRef, "")
			},
		},
	}

	mockUtils.AddFile(cnbutils.DefaultOrderPath, []byte(`[[order]]
	[[order.group]]
	  id = "buildpacks/java"
	  version = "1.8.0"
[[order]]
	[[order.group]]
	  id = "buildpacks/nodejs"
	  version = "1.6.0"`))

	t.Run("successfully loads baked in order.toml", func(t *testing.T) {
		order, err := cnbutils.CreateOrder(nil, nil, nil, "", mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, []cnbutils.OrderEntry{
			{
				Group: []cnbutils.BuildPackMetadata{
					{
						ID:      "buildpacks/java",
						Version: "1.8.0",
					},
				},
			},
			{
				Group: []cnbutils.BuildPackMetadata{
					{
						ID:      "buildpacks/nodejs",
						Version: "1.6.0",
					},
				},
			},
		}, order.Order)
	})

	t.Run("successfully loads baked in order.toml and adds pre/post buildpacks", func(t *testing.T) {
		order, err := cnbutils.CreateOrder(nil, []string{"pre-buildpack"}, []string{"post-buildpack"}, "", mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, []cnbutils.OrderEntry{
			{
				Group: []cnbutils.BuildPackMetadata{
					{
						ID:      "pre-testbuildpack",
						Version: "0.0.1",
					},
					{
						ID:      "buildpacks/java",
						Version: "1.8.0",
					},
					{
						ID:      "post-testbuildpack",
						Version: "0.0.1",
					},
				},
			},
			{
				Group: []cnbutils.BuildPackMetadata{
					{
						ID:      "pre-testbuildpack",
						Version: "0.0.1",
					},
					{
						ID:      "buildpacks/nodejs",
						Version: "1.6.0",
					},
					{
						ID:      "post-testbuildpack",
						Version: "0.0.1",
					},
				},
			},
		}, order.Order)
	})

	t.Run("successfully creates new order with custom buildpacks", func(t *testing.T) {
		order, err := cnbutils.CreateOrder([]string{"testbuildpack"}, nil, nil, "", mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, []cnbutils.OrderEntry{
			{
				Group: []cnbutils.BuildPackMetadata{
					{
						ID:      "testbuildpack",
						Version: "0.0.1",
					},
				},
			},
		}, order.Order)
	})

	t.Run("successfully creates new order with custom buildpacks and adds pre/post buildpacks", func(t *testing.T) {
		order, err := cnbutils.CreateOrder([]string{"testbuildpack"}, []string{"pre-buildpack"}, []string{"post-buildpack"}, "", mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, []cnbutils.OrderEntry{
			{
				Group: []cnbutils.BuildPackMetadata{
					{
						ID:      "pre-testbuildpack",
						Version: "0.0.1",
					},
					{
						ID:      "testbuildpack",
						Version: "0.0.1",
					},
					{
						ID:      "post-testbuildpack",
						Version: "0.0.1",
					},
				},
			},
		}, order.Order)
	})
}
