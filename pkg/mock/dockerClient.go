package mock

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/v1"
)

// DownloadMock .
type DownloadMock struct {
	FilePath    string
	ImageRef    string
	RegistryURL string

	ReturnImage v1.Image
	ReturnError string

	Stub func(imageRef, targetDir string) (v1.Image, error)
}

// DownloadImage .
func (c *DownloadMock) DownloadImage(imageRef, targetDir string) (v1.Image, error) {
	c.ImageRef = imageRef
	c.FilePath = targetDir

	if c.Stub != nil {
		return c.Stub(imageRef, targetDir)
	}

	if len(c.ReturnError) > 0 {
		return nil, fmt.Errorf(c.ReturnError)
	}
	return c.ReturnImage, nil
}

// DownloadImageContent .
func (c *DownloadMock) DownloadImageContent(imageRef, targetFile string) (v1.Image, error) {
	c.ImageRef = imageRef
	c.FilePath = targetFile

	if len(c.ReturnError) > 0 {
		return nil, fmt.Errorf(c.ReturnError)
	}
	return c.ReturnImage, nil
}
