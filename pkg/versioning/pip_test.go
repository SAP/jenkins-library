package versioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipInit(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		pip := Pip{}
		pip.init()
		assert.Equal(t, "VERSION", pip.path)
	})

	t.Run("no default", func(t *testing.T) {
		pip := Pip{path: "my/VERSION"}
		pip.init()
		assert.Equal(t, "my/VERSION", pip.path)
	})
}
