package mock

import (
	"context"
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

func (c *CraneMockUtils) CopyImage(_ context.Context, src, dest, platform string) error {
	return c.ErrCopyImage
}

func (c *CraneMockUtils) PushImage(_ context.Context, im v1.Image, dest, platform string) error {
	return c.ErrPushImage
}

func (c *CraneMockUtils) LoadImage(_ context.Context, src string) (v1.Image, error) {
	return nil, c.ErrLoadImage
}
