package docker

import (
	"context"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type CraneUtilsBundle struct{}

func (c *CraneUtilsBundle) CopyImage(ctx context.Context, src, dest string) error {
	return crane.Copy(src, dest, crane.WithContext(ctx))
}

func (c *CraneUtilsBundle) PushImage(ctx context.Context, im v1.Image, dest string) error {
	return crane.Push(im, dest, crane.WithContext(ctx))
}

func (c *CraneUtilsBundle) LoadImage(ctx context.Context, src string) (v1.Image, error) {
	return crane.Load(src, crane.WithContext(ctx))
}
