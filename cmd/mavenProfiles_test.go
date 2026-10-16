//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsProfile(t *testing.T) {
	t.Parallel()

	t.Run("nil slice returns false", func(t *testing.T) {
		assert.False(t, containsProfile(nil, "release.build"))
	})

	t.Run("empty slice returns false", func(t *testing.T) {
		assert.False(t, containsProfile([]string{}, "release.build"))
	})

	t.Run("target found returns true", func(t *testing.T) {
		assert.True(t, containsProfile([]string{"snapshot.build", "release.build", "custom"}, "release.build"))
	})

	t.Run("target not found returns false", func(t *testing.T) {
		assert.False(t, containsProfile([]string{"snapshot.build", "milestone.build"}, "release.build"))
	})
}
