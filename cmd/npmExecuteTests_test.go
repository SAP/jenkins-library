package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunNpmExecuteTests(t *testing.T) {
	t.Parallel()

	testCmd := NpmExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "npmExecuteTests", testCmd.Use, "command name incorrect")
}

func TestParseURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    []map[string]interface{}
		expected []vaultUrl
		wantErr  bool
	}{
		{
			name: "Valid URLs",
			input: []map[string]interface{}{
				{
					"url":      "http://example.com",
					"username": "user1",
					"password": "pass1",
				},
				{
					"url": "http://example2.com",
				},
			},
			expected: []vaultUrl{
				{
					URL:      "http://example.com",
					Username: "user1",
					Password: "pass1",
				},
				{
					URL: "http://example2.com",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid URL entry",
			input: []map[string]interface{}{
				{
					"username": "user1",
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Invalid URL field type",
			input: []map[string]interface{}{
				{
					"url": 123,
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Empty URLs",
			input:    []map[string]interface{}{},
			expected: []vaultUrl{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseURLs(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}
