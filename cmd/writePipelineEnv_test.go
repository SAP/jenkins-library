//go:build unit

package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/stretchr/testify/require"
)

func TestCleanJSONDataWritePipeline(t *testing.T) {
	t.Parallel()

	// Test valid UTF-8 (should be unchanged)
	validData := []byte(`{"git":{"commitMessage":"This is a valid commit message"}}`)
	result := piperenv.CleanJSONData(validData)
	require.Equal(t, validData, result)

	// Test emoji with valid UTF-8 (should be unchanged)
	emojiData := []byte(`{"git":{"commitMessage":"🚀 feat: add new feature"}}`)
	result = piperenv.CleanJSONData(emojiData)
	require.Equal(t, emojiData, result)

	// Test data with JSON control character (like \x16)
	invalidData := []byte("{\"git\":{\"commitMessage\":\"Test \x16 control char\"}}")
	result = piperenv.CleanJSONData(invalidData)
	require.NotEqual(t, invalidData, result)
	require.True(t, json.Valid(result), "Result should be valid JSON")

	// Verify we can parse the cleaned data as JSON
	var parsed map[string]interface{}
	err := json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	require.Contains(t, parsed, "git")

	// The control character should be escaped as unicode
	require.Contains(t, string(result), "\\u0016")
}

func TestParseInputWithInvalidUTF8(t *testing.T) {
	t.Parallel()

	// Test parsing JSON with control character
	invalidJSON := []byte("{\"git\":{\"commitMessage\":\"Test \x16 control char commit\"}}")

	// Should not fail due to control characters
	cpeMap, err := parseInput(invalidJSON)
	require.NoError(t, err)
	require.NotNil(t, cpeMap)

	// Verify we can access the git section
	if git, ok := cpeMap["git"]; ok {
		require.NotNil(t, git)
		if gitMap, ok := git.(map[string]interface{}); ok {
			commitMsg, exists := gitMap["commitMessage"]
			require.True(t, exists)
			require.IsType(t, "", commitMsg)
			// The message should be present and non-empty
			require.NotEmpty(t, commitMsg)
		}
	}
}

func TestParseInputWithValidEmoji(t *testing.T) {
	t.Parallel()

	// Test parsing JSON with valid emoji
	validJSON := []byte(`{"git":{"commitMessage":"🚀 feat: add new feature"}}`)

	cpeMap, err := parseInput(validJSON)
	require.NoError(t, err)
	require.NotNil(t, cpeMap)

	// Verify emoji is preserved
	if git, ok := cpeMap["git"]; ok {
		if gitMap, ok := git.(map[string]interface{}); ok {
			commitMsg, exists := gitMap["commitMessage"]
			require.True(t, exists)
			require.Equal(t, "🚀 feat: add new feature", commitMsg)
		}
	}
}
