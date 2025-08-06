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
