package mock

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// DownloadMock .
type DownloadMock struct {
	FilePath       string
	ImageRef       string
	RemoteImageRef string
	RegistryURL    string

	ReturnImage     v1.Image
	RemoteImageInfo v1.Image
	ReturnError     string

	Stub             func(imageRef, targetDir string) (v1.Image, error)
	ImageContentStub func(imageRef, targetFile string) (v1.Image, error)
	ImageInfoStub    func(imageRef string) (v1.Image, error)
}

// DownloadImage .
func (c *DownloadMock) DownloadImage(imageRef, targetDir string) (v1.Image, error) {
	c.ImageRef = imageRef
	c.FilePath = targetDir

	if c.Stub != nil {
		return c.Stub(imageRef, targetDir)
	}

	if len(c.ReturnError) > 0 {
		return nil, fmt.Errorf("%s", c.ReturnError)
	}
	return c.ReturnImage, nil
}

// DownloadImageContent .
func (c *DownloadMock) DownloadImageContent(imageRef, targetFile string) (v1.Image, error) {
	c.ImageRef = imageRef
	c.FilePath = targetFile

	if c.ImageContentStub != nil {
		return c.ImageContentStub(imageRef, targetFile)
	}

	if len(c.ReturnError) > 0 {
		return nil, fmt.Errorf("%s", c.ReturnError)
	}
	return c.ReturnImage, nil
}

// GetRemoteImageInfo .
func (c *DownloadMock) GetRemoteImageInfo(imageRef string) (v1.Image, error) {
	c.RemoteImageRef = imageRef

	if c.ImageInfoStub != nil {
		return c.ImageInfoStub(imageRef)
	}

	if len(c.ReturnError) > 0 {
		return nil, fmt.Errorf("%s", c.ReturnError)
	}

	return c.RemoteImageInfo, nil
}
