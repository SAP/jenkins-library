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

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		input    map[string]interface{}
		expected *parsedMetadata
		name     string
		wantErr  bool
	}{
		{
			name: "Valid metadata with URLs",
			input: map[string]interface{}{
				"urls": []interface{}{
					map[string]interface{}{
						"url":      "http://example.com",
						"username": "user1",
						"password": "pass1",
					},
					map[string]interface{}{
						"url": "http://example2.com",
					},
				},
				"username": "globalUser",
				"password": "globalPass",
			},
			expected: &parsedMetadata{
				GlobalUsername: "globalUser",
				GlobalPassword: "globalPass",
				URLs: []appUrl{
					{
						URL:      "http://example.com",
						Username: "user1",
						Password: "pass1",
					},
					{
						URL: "http://example2.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid metadata without URLs",
			input: map[string]interface{}{
				"username": "globalUser",
				"password": "globalPass",
			},
			expected: &parsedMetadata{
				GlobalUsername: "globalUser",
				GlobalPassword: "globalPass",
				URLs:           []appUrl{},
			},
			wantErr: false,
		},
		{
			name: "Invalid URL entry",
			input: map[string]interface{}{
				"urls": []interface{}{
					"invalidEntry",
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Invalid URL field type",
			input: map[string]interface{}{
				"urls": []interface{}{
					map[string]interface{}{
						"url": 123,
					},
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:  "Empty metadata",
			input: map[string]interface{}{},
			expected: &parsedMetadata{
				URLs: []appUrl{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMetadata(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}
