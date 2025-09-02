package feature

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFeatureEnabled(t *testing.T) {
	t.Run("", func(t *testing.T) {
		assert.False(t, IsFeatureEnabled("newFeature"))

		// defer resetEnv(os.Environ())
		os.Setenv(prefix+"newFeature", "true")
		defer os.Setenv(prefix+"newFeature", "")

		assert.True(t, IsFeatureEnabled("newFeature"))
	})
}
