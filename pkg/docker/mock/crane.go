package mock

import v1 "github.com/google/go-containerregistry/pkg/v1"

type CraneMockUtils struct {
	errCopyImage, errPushImage, errLoadImage error
}

func (c *CraneMockUtils) CopyImage(src string, dest string) error {
	return c.errCopyImage
}

func (c *CraneMockUtils) PushImage(im v1.Image, dest string) error {
	return c.errPushImage
}

func (c *CraneMockUtils) LoadImage(src string) (v1.Image, error) {
	return nil, c.errLoadImage
}
