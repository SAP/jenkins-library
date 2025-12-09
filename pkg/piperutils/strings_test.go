//go:build unit
// +build unit

package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitle(t *testing.T) {
	assert.Equal(t, "TEST", Title("tEST"))
	assert.Equal(t, "Test", Title("test"))
	assert.Equal(t, "TEST", Title("TEST"))
	assert.Equal(t, "Test", Title("Test"))
	assert.Equal(t, "TEST1 Test2 TEsT3 Test4", Title("TEST1 test2 tEsT3 Test4"))
}

func TestStringWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		want         string
	}{
		{"Non-empty input", "foo", "bar", "foo"},
		{"Input with spaces", "  foo  ", "bar", "foo"},
		{"Empty input", "", "bar", "bar"},
		{"Whitespace input", "   ", "bar", "bar"},
		{"Empty default", "", "", ""},
		{"Whitespace input, empty default", "   ", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringWithDefault(tt.input, tt.defaultValue)
			if got != tt.want {
				t.Errorf("StringWithDefault(%q, %q) = %q, want %q", tt.input, tt.defaultValue, got, tt.want)
			}
		})
	}
}

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
