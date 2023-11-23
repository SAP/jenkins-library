package docker

import (
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type CraneUtilsBundle struct{}

func (c *CraneUtilsBundle) CopyImage(src, dest string) error {
	return crane.Copy(src, dest)
}

func (c *CraneUtilsBundle) PushImage(im v1.Image, dest string) error {
	return crane.Push(im, dest)
}

func (c *CraneUtilsBundle) LoadImage(src string) (v1.Image, error) {
	return crane.Load(src)
}
