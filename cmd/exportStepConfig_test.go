//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name:     "Double dash with value",
			args:     []string{"--flag1", "value1s"},
			expected: map[string]string{"flag1": "value1"},
		},
		{
			name:     "Double dash with equals",
			args:     []string{"--flag2=value2"},
			expected: map[string]string{"flag2": "value2"},
		},
		{
			name:     "Shorthand with value",
			args:     []string{"-f1", "value3"},
			expected: map[string]string{"f1": "value3"},
		},
		{
			name:     "Boolean flag",
			args:     []string{"--verbose"},
			expected: map[string]string{"verbose": "true"},
		},
		{
			name:     "Mixed flags",
			args:     []string{"--flag1", "value1", "--flag2=value2", "-f1", "value3", "--verbose"},
			expected: map[string]string{"flag1": "value1", "flag2": "value2", "f1": "value3", "verbose": "true"},
		},
		{
			name:     "No flags",
			args:     []string{"some", "random", "args"},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArgs(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}
