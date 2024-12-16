//go:build unit

package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_escapeValuesForCLI(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		input    string
		expected string
	}{
		{
			name:     "Windows password without quotes",
			os:       "windows",
			input:    `mypassword`,
			expected: `'mypassword'`,
		},
		{
			name:     "Windows password with quotes",
			os:       "windows",
			input:    `my\"password`,
			expected: `'my\"password'`,
		},
		{
			name:     "Non-Windows password without single quotes",
			os:       "linux",
			input:    "mypassword",
			expected: "'mypassword'",
		},
		{
			name:     "Non-Windows password with single quotes",
			os:       "darwin",
			input:    `my'password`,
			expected: `'my'\''password'`,
		},
		{
			name:     "Linux password with all special characters",
			os:       "linux",
			input:    "~!@#$%^&*()_+{`}|:\"<>?-=[]\\;',./",
			expected: "'~!@#$%^&*()_+{`}|:\"<>?-=[]\\;'\\'',./'",
		},
		{
			name:     "Windows password with all special characters",
			os:       "windows",
			input:    "~!@#$%^&*()_+{`}|:\"<>?-=[]\\;',./",
			expected: "'~!@#$%^&*()_+{`}|:\"<>?-=[]\\;'',./'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeValuesForCLI(tt.input, func() string { return tt.os })
			assert.Equal(t, tt.expected, result, fmt.Sprintf("Failed for OS: %s and password: %s", tt.os, tt.input))
		})
	}
}
