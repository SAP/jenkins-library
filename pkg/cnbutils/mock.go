//go:build !release
// +build !release

package cnbutils

import (
	"io"
	"path/filepath"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	fakeImage "github.com/google/go-containerregistry/pkg/v1/fake"
)

type MockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (c *MockUtils) GetFileUtils() piperutils.FileUtils {
	return c.FilesMock
}

func (c *MockUtils) DownloadImageToPath(bpack, filePath string) (pkgutil.Image, error) {
	fakeImage := fakeImage.FakeImage{}
	fakeImage.ConfigFileReturns(&v1.ConfigFile{
		Config: v1.Config{
			Labels: map[string]string{
				"io.buildpacks.buildpackage.metadata": "{\"id\": \"testbuildpack\", \"version\": \"0.0.1\"}",
			},
		},
	}, nil)

	c.AddDir(filepath.Join(filePath, "cnb/buildpacks", bpack))
	c.AddDir(filepath.Join(filePath, "cnb/buildpacks", bpack, "0.0.1"))
	img := pkgutil.Image{
		Image: &fakeImage,
	}
	return img, nil
}

func (c *MockUtils) GetImageSource() (string, error) {
	return "imageSource", nil
}

func (c *MockUtils) TarImage(writer io.Writer, image pkgutil.Image) error {
	return nil
}
