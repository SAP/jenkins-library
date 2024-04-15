package codeql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePatterns(t *testing.T) {
	t.Parallel()

	t.Run("Empty input", func(t *testing.T) {
		input := []string{}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.Empty(t, patterns)
	})

	t.Run("One pattern to exclude", func(t *testing.T) {
		input := []string{"-file_pattern"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "file_pattern", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
	})

	t.Run("One pattern to include", func(t *testing.T) {
		input := []string{"+file_pattern"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "file_pattern", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.True(t, patterns[0].sign)
	})

	t.Run("One pattern without sign", func(t *testing.T) {
		input := []string{"file_pattern"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "file_pattern", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.True(t, patterns[0].sign)
	})

	t.Run("Several patterns to exclude", func(t *testing.T) {
		input := []string{"-file_pattern_1", "-file_pattern_2"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "file_pattern_1", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
		assert.Equal(t, "file_pattern_2", patterns[1].filePattern)
		assert.Equal(t, "**", patterns[1].rulePattern)
		assert.False(t, patterns[1].sign)
	})

	t.Run("Several patterns to include", func(t *testing.T) {
		input := []string{"+file_pattern_1", "file_pattern_2"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "file_pattern_1", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.True(t, patterns[0].sign)
		assert.Equal(t, "file_pattern_2", patterns[1].filePattern)
		assert.Equal(t, "**", patterns[1].rulePattern)
		assert.True(t, patterns[1].sign)
	})

	t.Run("One pattern to exclude, one pattern to include", func(t *testing.T) {
		input := []string{"-file_pattern_1", "+file_pattern_2"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "file_pattern_1", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
		assert.Equal(t, "file_pattern_2", patterns[1].filePattern)
		assert.Equal(t, "**", patterns[1].rulePattern)
		assert.True(t, patterns[1].sign)
	})

	t.Run("Several patterns to exclude and include", func(t *testing.T) {
		input := []string{"-file_pattern_1", "+file_pattern_2", "-file_pattern_3", "file_pattern_4"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 4, len(patterns))
		assert.Equal(t, "file_pattern_1", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
		assert.Equal(t, "file_pattern_2", patterns[1].filePattern)
		assert.Equal(t, "**", patterns[1].rulePattern)
		assert.True(t, patterns[1].sign)
		assert.Equal(t, "file_pattern_3", patterns[2].filePattern)
		assert.Equal(t, "**", patterns[2].rulePattern)
		assert.False(t, patterns[2].sign)
		assert.Equal(t, "file_pattern_4", patterns[3].filePattern)
		assert.Equal(t, "**", patterns[3].rulePattern)
		assert.True(t, patterns[3].sign)
	})

	t.Run("Patterns with spaces", func(t *testing.T) {
		input := []string{"-file pattern 1", "-file pattern 2"}
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "file pattern 1", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
		assert.Equal(t, "file pattern 2", patterns[1].filePattern)
		assert.Equal(t, "**", patterns[1].rulePattern)
		assert.False(t, patterns[1].sign)
	})

	t.Run("Patterns with slashes", func(t *testing.T) {
		input := []string{"-file/pattern/1", "-file\\\\pattern\\\\2"} // -file\pattern\2
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "file/pattern/1", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
		assert.Equal(t, "file\\pattern\\2", patterns[1].filePattern)
		assert.Equal(t, "**", patterns[1].rulePattern)
		assert.False(t, patterns[1].sign)
	})

	t.Run("Invalid pattern", func(t *testing.T) {
		input := []string{"file :pattern:rule"}
		_, err := ParsePatterns(input)
		assert.Error(t, err)
	})
}

func TestParsePattern(t *testing.T) {
	t.Parallel()

	t.Run("Empty string", func(t *testing.T) {
		input := ""
		pattern, err := parsePattern(input)
		assert.NoError(t, err)
		assert.True(t, pattern.sign)
		assert.Equal(t, "", pattern.filePattern)
		assert.Equal(t, "**", pattern.rulePattern)
	})

	t.Run("Include files, no rules", func(t *testing.T) {
		input := "+file_pattern"
		pattern, err := parsePattern(input)
		assert.NoError(t, err)
		assert.True(t, pattern.sign)
		assert.Equal(t, "file_pattern", pattern.filePattern)
		assert.Equal(t, "**", pattern.rulePattern)
	})

	t.Run("Exclude files, no rules", func(t *testing.T) {
		input := "-file_pattern"
		pattern, err := parsePattern(input)
		assert.NoError(t, err)
		assert.False(t, pattern.sign)
		assert.Equal(t, "file_pattern", pattern.filePattern)
		assert.Equal(t, "**", pattern.rulePattern)
	})

	t.Run("Include files with rule", func(t *testing.T) {
		input := "+file_pattern:rule"
		pattern, err := parsePattern(input)
		assert.NoError(t, err)
		assert.True(t, pattern.sign)
		assert.Equal(t, "file_pattern", pattern.filePattern)
		assert.Equal(t, "rule", pattern.rulePattern)
	})

	t.Run("Exclude files with rule", func(t *testing.T) {
		input := "-file_pattern:rule"
		pattern, err := parsePattern(input)
		assert.NoError(t, err)
		assert.False(t, pattern.sign)
		assert.Equal(t, "file_pattern", pattern.filePattern)
		assert.Equal(t, "rule", pattern.rulePattern)
	})

	t.Run("Pattern without sign", func(t *testing.T) {
		input := "file_pattern:rule"
		pattern, err := parsePattern(input)
		assert.NoError(t, err)
		assert.True(t, pattern.sign)
		assert.Equal(t, "file_pattern", pattern.filePattern)
		assert.Equal(t, "rule", pattern.rulePattern)
	})

	t.Run("Pattern with escape character", func(t *testing.T) {
		input := "\\+file_pattern:\\:rule"
		pattern, err := parsePattern(input)
		assert.NoError(t, err)
		assert.True(t, pattern.sign)
		assert.Equal(t, "+file_pattern", pattern.filePattern)
		assert.Equal(t, ":rule", pattern.rulePattern)
	})

	t.Run("Pattern with duplicated separator", func(t *testing.T) {
		input := "file_pattern::rule"
		_, err := parsePattern(input)
		assert.Error(t, err)
	})
}

func TestGetSignAndTrimPattern(t *testing.T) {
	t.Parallel()

	t.Run("Pattern to include with sign", func(t *testing.T) {
		input := "+pattern"
		include, pattern := getSignAndTrimPattern(input)
		assert.True(t, include)
		assert.Equal(t, "pattern", pattern)
	})

	t.Run("Pattern to include without sign", func(t *testing.T) {
		input := "pattern"
		include, pattern := getSignAndTrimPattern(input)
		assert.True(t, include)
		assert.Equal(t, "pattern", pattern)
	})

	t.Run("Pattern to include with sign", func(t *testing.T) {
		input := "-pattern"
		include, pattern := getSignAndTrimPattern(input)
		assert.False(t, include)
		assert.Equal(t, "pattern", pattern)
	})

	t.Run("Empty input", func(t *testing.T) {
		input := ""
		include, pattern := getSignAndTrimPattern(input)
		assert.True(t, include)
		assert.Equal(t, "", pattern)
	})
}

func TestSeparateFileAndRulePattern(t *testing.T) {
	t.Parallel()

	t.Run("File pattern without rule pattern", func(t *testing.T) {
		input := "file_pattern"
		filePattern, rulePattern, err := separateFileAndRulePattern(input)
		assert.NoError(t, err)
		assert.Equal(t, "file_pattern", filePattern)
		assert.Equal(t, "", rulePattern)
	})

	t.Run("File pattern with rule pattern", func(t *testing.T) {
		input := "file_pattern:rule"
		filePattern, rulePattern, err := separateFileAndRulePattern(input)
		assert.NoError(t, err)
		assert.Equal(t, "file_pattern", filePattern)
		assert.Equal(t, "rule", rulePattern)
	})

	t.Run("Escaped separator", func(t *testing.T) {
		input := "file\\:pattern:rule"
		filePattern, rulePattern, err := separateFileAndRulePattern(input)
		assert.NoError(t, err)
		assert.Equal(t, "file:pattern", filePattern)
		assert.Equal(t, "rule", rulePattern)
	})

	t.Run("Escaped escape character", func(t *testing.T) {
		input := "file_pattern\\\\:rule"
		filePattern, rulePattern, err := separateFileAndRulePattern(input)
		assert.NoError(t, err)
		assert.Equal(t, "file_pattern\\", filePattern)
		assert.Equal(t, "rule", rulePattern)
	})

	t.Run("Multiple separators", func(t *testing.T) {
		input := "file:pattern:rule"
		_, _, err := separateFileAndRulePattern(input)
		assert.Error(t, err)
	})

	t.Run("Empty string", func(t *testing.T) {
		input := ""
		filePattern, rulePattern, err := separateFileAndRulePattern(input)
		assert.NoError(t, err)
		assert.Equal(t, "", filePattern)
		assert.Equal(t, "", rulePattern)
	})

	t.Run("Separator at first position", func(t *testing.T) {
		input := ":rule"
		filePattern, rulePattern, err := separateFileAndRulePattern(input)
		assert.NoError(t, err)
		assert.Equal(t, "", filePattern)
		assert.Equal(t, "rule", rulePattern)
	})

	t.Run("Separator at last position", func(t *testing.T) {
		input := "file_pattern:"
		filePattern, rulePattern, err := separateFileAndRulePattern(input)
		assert.NoError(t, err)
		assert.Equal(t, "file_pattern", filePattern)
		assert.Equal(t, "", rulePattern)
	})
}

func TestMatchPathAndRule(t *testing.T) {
	t.Parallel()

	t.Run("Single pattern match, file and rule will be included to results", func(t *testing.T) {
		path := "path/to/src/file"
		ruleId := "rule"
		patterns := []*Pattern{
			{
				sign:        true,
				filePattern: "**/file",
				rulePattern: "rule",
			},
		}
		include, err := matchPathAndRule(path, ruleId, patterns)
		assert.NoError(t, err)
		assert.True(t, include)
	})

	t.Run("Single pattern match, file and rule will be excluded from results", func(t *testing.T) {
		path := "path/to/src/file"
		ruleId := "rule"
		patterns := []*Pattern{
			{
				sign:        false,
				filePattern: "**/file",
				rulePattern: "rule",
			},
		}
		include, err := matchPathAndRule(path, ruleId, patterns)
		assert.NoError(t, err)
		assert.False(t, include)
	})

	t.Run("Multiple patterns match, file and rule will be included to results", func(t *testing.T) {
		path := "path/to/src/file"
		ruleId := "rule1"
		patterns := []*Pattern{
			{
				sign:        true,
				filePattern: "**/file",
				rulePattern: "rule1",
			},
			{
				sign:        false,
				rulePattern: "rule2",
				filePattern: "**/file.go",
			},
		}
		include, err := matchPathAndRule(path, ruleId, patterns)
		assert.NoError(t, err)
		assert.True(t, include)
	})

	t.Run("Multiple patterns match, file and rule will be excluded from results", func(t *testing.T) {
		path := "path/to/src/file"
		ruleId := "rule1"
		patterns := []*Pattern{
			{
				sign:        true,
				filePattern: "**/**",
				rulePattern: "**",
			},
			{
				sign:        false,
				rulePattern: "**",
				filePattern: "**/file",
			},
		}
		include, err := matchPathAndRule(path, ruleId, patterns)
		assert.NoError(t, err)
		assert.False(t, include)
	})

	t.Run("No matches, path and rule will be included to results", func(t *testing.T) {
		path := "path/to/src/file"
		ruleId := "rule"
		patterns := []*Pattern{
			{
				sign:        false,
				filePattern: "**/file.??",
				rulePattern: "rule1",
			},
			{
				sign:        false,
				rulePattern: "rule2",
				filePattern: "path/*",
			},
		}
		include, err := matchPathAndRule(path, ruleId, patterns)
		assert.NoError(t, err)
		assert.True(t, include)
	})

	t.Run("Invalid pattern", func(t *testing.T) {
		path := "path/to/src/file"
		ruleId := "rule"
		patterns := []*Pattern{
			{
				sign:        false,
				filePattern: "path/[",
				rulePattern: "rule*",
			},
		}
		_, err := matchPathAndRule(path, ruleId, patterns)
		assert.Error(t, err)
	})

	t.Run("Empty path", func(t *testing.T) {
		path := ""
		ruleId := "rule"
		patterns := []*Pattern{
			{
				sign:        false,
				filePattern: "*",
				rulePattern: "rule*",
			},
		}
		include, err := matchPathAndRule(path, ruleId, patterns)
		assert.NoError(t, err)
		assert.False(t, include)
	})
}

func TestFilterSarif(t *testing.T) {
	t.Parallel()

	t.Run("Nothing to exclude", func(t *testing.T) {
		input := map[string]interface{}{
			"runs": []interface{}{
				map[string]interface{}{
					"results": []interface{}{
						map[string]interface{}{
							"ruleId": "rule1",
							"locations": []interface{}{
								map[string]interface{}{
									"physicalLocation": map[string]interface{}{
										"artifactLocation": map[string]interface{}{
											"uri": "myapp/modules/main.go",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		patterns := []*Pattern{
			{
				sign:        false,
				filePattern: "**/src/**",
				rulePattern: "**",
			},
		}
		filteredSarif, err := FilterSarif(input, patterns)
		assert.NoError(t, err)
		results, ok := filteredSarif["runs"].([]interface{})[0].(map[string]interface{})["results"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 1, len(results))
	})

	t.Run("Exclude single result", func(t *testing.T) {
		input := map[string]interface{}{
			"runs": []interface{}{
				map[string]interface{}{
					"results": []interface{}{
						map[string]interface{}{
							"ruleId": "rule1",
							"locations": []interface{}{
								map[string]interface{}{
									"physicalLocation": map[string]interface{}{
										"artifactLocation": map[string]interface{}{
											"uri": "myapp/modules/main.go",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		patterns := []*Pattern{
			{
				sign:        false,
				filePattern: "myapp/**",
				rulePattern: "**",
			},
		}
		filteredSarif, err := FilterSarif(input, patterns)
		assert.NoError(t, err)
		results, ok := filteredSarif["runs"].([]interface{})[0].(map[string]interface{})["results"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 0, len(results))
	})
}
