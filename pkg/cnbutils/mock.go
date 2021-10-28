//go:build !release
// +build !release

package cnbutils

import (
	"io"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	fakeImage "github.com/google/go-containerregistry/pkg/v1/fake"
)

type MockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*DockerMock
}

func (c *MockUtils) GetDockerClient() docker.Download {
	return c.DockerMock
}

func (c *MockUtils) GetFileUtils() piperutils.FileUtils {
	return c.FilesMock
}

type DockerMock struct{}

func (d *DockerMock) DownloadImageToPath(_, filePath string) (pkgutil.Image, error) {

	fakeImage := fakeImage.FakeImage{}
	fakeImage.ConfigFileReturns(&v1.ConfigFile{
		Config: v1.Config{
			Labels: map[string]string{
				"io.buildpacks.buildpackage.metadata": "{\"id\": \"testbuildpack\", \"version\": \"0.0.1\"}",
			},
		},
	}, nil)
	img := pkgutil.Image{
		Image: &fakeImage,
	}
	return img, nil
}

func (d *DockerMock) GetImageSource() (string, error) {
	return "imageSource", nil
}

func (d *DockerMock) TarImage(writer io.Writer, image pkgutil.Image) error {
	return nil
}
