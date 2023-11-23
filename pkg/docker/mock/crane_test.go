package mock

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	src = "source"
	dst = "destination"
)

func TestCopyImage(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		crane := CraneMockUtils{}
		err := crane.CopyImage(src, dst)
		assert.NoError(t, err)
	})
	t.Run("bad case", func(t *testing.T) {
		crane := CraneMockUtils{errCopyImage: errors.New("copy image err")}
		err := crane.CopyImage(src, dst)
		assert.EqualError(t, err, "copy image err")
	})
}

func TestPushImage(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		crane := CraneMockUtils{}
		err := crane.PushImage(nil, dst)
		assert.NoError(t, err)
	})
	t.Run("bad case", func(t *testing.T) {
		crane := CraneMockUtils{errCopyImage: errors.New("push image err")}
		err := crane.PushImage(nil, dst)
		assert.EqualError(t, err, "push image err")
	})
}

func TestLoadImage(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		crane := CraneMockUtils{}
		_, err := crane.LoadImage(src)
		assert.NoError(t, err)
	})
	t.Run("bad case", func(t *testing.T) {
		crane := CraneMockUtils{errCopyImage: errors.New("load image err")}
		_, err := crane.LoadImage(src)
		assert.EqualError(t, err, "load image err")
	})
}
