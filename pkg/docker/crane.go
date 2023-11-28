package docker

import (
	"context"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type CraneUtilsBundle struct{}

func (c *CraneUtilsBundle) CopyImage(ctx context.Context, src, dest, platform string) error {
	p, err := v1.ParsePlatform(platform)
	if err != nil {
		return err
	}
	return crane.Copy(src, dest, crane.WithContext(ctx), crane.WithPlatform(p))
}

func (c *CraneUtilsBundle) PushImage(ctx context.Context, im v1.Image, dest, platform string) error {
	p, err := v1.ParsePlatform(platform)
	if err != nil {
		return err
	}
	return crane.Push(im, dest, crane.WithContext(ctx), crane.WithPlatform(p))
}

func (c *CraneUtilsBundle) LoadImage(ctx context.Context, src string) (v1.Image, error) {
	return crane.Load(src, crane.WithContext(ctx))
}
