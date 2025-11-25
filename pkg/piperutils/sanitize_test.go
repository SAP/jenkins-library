package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizePath(t *testing.T) {
	t.Run("URL with query parameters", func(t *testing.T) {
		input := "https://example.com/dir/file.txt?param=value"
		expected := "https://example.com/dir/file.txt"
		assert.Equal(t, expected, SanitizePath(input))
	})

	t.Run("File path with query parameters", func(t *testing.T) {
		input := "invalid-url/file.txt?param=value"
		expected := "invalid-url/file.txt"
		assert.Equal(t, expected, SanitizePath(input))
	})

	t.Run("Path without query parameters", func(t *testing.T) {
		input := "/dir/file.txt"
		expected := "/dir/file.txt"
		assert.Equal(t, expected, SanitizePath(input))
	})

	t.Run("Multiple query parameters", func(t *testing.T) {
		input := "https://api.github.com/script.sh?token=abc&param=xyz"
		expected := "https://api.github.com/script.sh"
		assert.Equal(t, expected, SanitizePath(input))
	})

	t.Run("Local path with query", func(t *testing.T) {
		input := "./script.sh?arg=value"
		expected := "./script.sh"
		assert.Equal(t, expected, SanitizePath(input))
	})

	t.Run("Empty string", func(t *testing.T) {
		input := ""
		expected := ""
		assert.Equal(t, expected, SanitizePath(input))
	})

	t.Run("Only query parameter", func(t *testing.T) {
		input := "?param=value"
		expected := ""
		assert.Equal(t, expected, SanitizePath(input))
	})
}
