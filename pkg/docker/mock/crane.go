package mock

import (
	"errors"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

var (
	ErrCopyImage = errors.New("copy image err")
	ErrPushImage = errors.New("push image err")
	ErrLoadImage = errors.New("load image err")
)

type CraneMockUtils struct {
	ErrCopyImage, ErrPushImage, ErrLoadImage error
}

func (c *CraneMockUtils) CopyImage(src string, dest string) error {
	return c.ErrCopyImage
}

func (c *CraneMockUtils) PushImage(im v1.Image, dest string) error {
	return c.ErrPushImage
}

func (c *CraneMockUtils) LoadImage(src string) (v1.Image, error) {
	return nil, c.ErrLoadImage
}
