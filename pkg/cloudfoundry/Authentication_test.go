//go:build unit
// +build unit

package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreparePasswordForCLI(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		password string
		expected string
	}{
		{
			name:     "Windows password with no quotes",
			os:       "windows",
			password: "mypassword",
			expected: "\"mypassword\"",
		},
		{
			name:     "Windows password with quotes",
			os:       "windows",
			password: "my\"password",
			expected: "\"my\\\"password\"",
		},
		{
			name:     "Non-Windows password with no single quotes",
			os:       "linux",
			password: "mypassword",
			expected: "'mypassword'",
		},
		{
			name:     "Non-Windows password with single quotes",
			os:       "darwin",
			password: "my'password",
			expected: "'my'\\''password'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//originalGOOS := getGOOS
			getGOOS := func() string { return tt.os } // Mock the OS
			//defer func() { getGOOS = originalGOOS }() // Restore the original OS after the test

			result := preparePasswordForCLI(tt.password, getGOOS)
			assert.Equal(t, tt.expected, result, fmt.Sprintf("Failed for OS: %s and password: %s", tt.os, tt.password))
		})
	}
}
