//go:build !release
// +build !release

package cnbutils

import (
	"fmt"
	"path/filepath"

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

func (c *MockUtils) DownloadImageContent(bpack, targetDir string) (v1.Image, error) {
	fakeImage := fakeImage.FakeImage{}
	fakeImage.ConfigFileReturns(&v1.ConfigFile{
		Config: v1.Config{
			Labels: map[string]string{
				"io.buildpacks.buildpackage.metadata": "{\"id\": \"testbuildpack\", \"version\": \"0.0.1\"}",
			},
		},
	}, nil)

	c.AddDir(filepath.Join(targetDir, "cnb/buildpacks", bpack))
	c.AddDir(filepath.Join(targetDir, "cnb/buildpacks", bpack, "0.0.1"))
	return &fakeImage, nil
}

func (c *MockUtils) DownloadImage(src, dst string) (v1.Image, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *MockUtils) GetImageSource() (string, error) {
	return "imageSource", nil
}
