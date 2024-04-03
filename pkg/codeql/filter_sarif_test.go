package codeql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePatterns(t *testing.T) {
	t.Parallel()

	t.Run("Empty input", func(t *testing.T) {
		input := ""
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.Empty(t, patterns)
	})

	t.Run("One pattern to exclude", func(t *testing.T) {
		input := "-file_pattern"
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "file_pattern", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
	})

	t.Run("One pattern to include", func(t *testing.T) {
		input := "+file_pattern"
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "file_pattern", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.True(t, patterns[0].sign)
	})

	t.Run("One pattern without sign", func(t *testing.T) {
		input := "file_pattern"
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "file_pattern", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.True(t, patterns[0].sign)
	})

	t.Run("Several patterns to exclude", func(t *testing.T) {
		input := "-file_pattern_1 -file_pattern_2"
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
		input := "+file_pattern_1 file_pattern_2"
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
		input := "-file_pattern_1 +file_pattern_2"
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
		input := "-file_pattern_1 +file_pattern_2 -file_pattern_3 file_pattern_4"
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

	t.Run("Pattern with spaces", func(t *testing.T) {
		input := "-file\\ pattern"
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "file\\ pattern", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
	})

	t.Run("Patterns with spaces", func(t *testing.T) {
		input := "-file\\ pattern\\ 1 -file\\ pattern\\ 2"
		patterns, err := ParsePatterns(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "file\\ pattern\\ 1", patterns[0].filePattern)
		assert.Equal(t, "**", patterns[0].rulePattern)
		assert.False(t, patterns[0].sign)
		assert.Equal(t, "file\\ pattern\\ 2", patterns[1].filePattern)
		assert.Equal(t, "**", patterns[1].rulePattern)
		assert.False(t, patterns[1].sign)
	})

	t.Run("Invalid pattern", func(t *testing.T) {
		input := "file\\ :pattern:rule"
		_, err := ParsePatterns(input)
		assert.Error(t, err)
	})
}

func TestSplit(t *testing.T) {
	t.Parallel()

	t.Run("Empty string", func(t *testing.T) {
		input := ""
		patterns := split(input)
		assert.Equal(t, 0, len(patterns))
	})

	t.Run("Patterns with spaces", func(t *testing.T) {
		input := "-file\\ pattern -file\\ pattern\\ 2"
		patterns := split(input)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "-file\\ pattern", patterns[0])
		assert.Equal(t, "-file\\ pattern\\ 2", patterns[1])
	})

	t.Run("First pattern with space", func(t *testing.T) {
		input := "-file\\ pattern -file_pattern"
		patterns := split(input)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "-file\\ pattern", patterns[0])
		assert.Equal(t, "-file_pattern", patterns[1])
	})

	t.Run("Second pattern with space", func(t *testing.T) {
		input := "-file_pattern -file\\ pattern"
		patterns := split(input)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "-file_pattern", patterns[0])
		assert.Equal(t, "-file\\ pattern", patterns[1])
	})

	t.Run("Patterns without spaces", func(t *testing.T) {
		input := "-file_pattern_1 file_pattern_2 +file_pattern_3"
		patterns := split(input)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 3, len(patterns))
		assert.Equal(t, "-file_pattern_1", patterns[0])
		assert.Equal(t, "file_pattern_2", patterns[1])
		assert.Equal(t, "+file_pattern_3", patterns[2])
	})

	t.Run("Second pattern with escape symbol", func(t *testing.T) {
		input := "-file_pattern_1 -file_pattern_2\\"
		patterns := split(input)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 2, len(patterns))
		assert.Equal(t, "-file_pattern_1", patterns[0])
		assert.Equal(t, "-file_pattern_2\\", patterns[1])
	})

	t.Run("First pattern with escape symbol", func(t *testing.T) {
		input := "-file_pattern_1\\ -file_pattern_2"
		patterns := split(input)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "-file_pattern_1\\ -file_pattern_2", patterns[0])
	})

	t.Run("Pattern with several spaces", func(t *testing.T) {
		input := "-file\\ pattern\\ 1"
		patterns := split(input)
		assert.NotEmpty(t, patterns)
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "-file\\ pattern\\ 1", patterns[0])
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

func TestMatchComponent(t *testing.T) {
	t.Parallel()

	t.Run("Empty inputs", func(t *testing.T) {
		fileName := ""
		pattern := ""
		assert.True(t, matchComponent(pattern, fileName))
	})

	t.Run("Empty pattern, non-empty filename", func(t *testing.T) {
		fileName := "file"
		pattern := ""
		assert.False(t, matchComponent(pattern, fileName))
	})

	t.Run("Empty filename, pattern is *", func(t *testing.T) {
		fileName := ""
		pattern := "*"
		assert.True(t, matchComponent(pattern, fileName))
	})

	t.Run("Non-empty pattern, empty filename", func(t *testing.T) {
		fileName := ""
		pattern := "pattern"
		assert.False(t, matchComponent(pattern, fileName))
	})

	t.Run("Pattern starts with *, non-empty filename", func(t *testing.T) {
		fileName := "file"
		pattern := "*"
		assert.True(t, matchComponent(pattern, fileName))
	})

	t.Run("Pattern starts with ? matches non-empty filename", func(t *testing.T) {
		fileName := "file"
		pattern := "????"
		assert.True(t, matchComponent(pattern, fileName))
	})

	t.Run("Pattern starts with ? doesn't match non-empty filename", func(t *testing.T) {
		fileName := "file"
		pattern := "?"
		assert.False(t, matchComponent(pattern, fileName))
	})

	t.Run("Escape symbol in pattern, matches filename", func(t *testing.T) {
		fileName := "file"
		pattern := "\\file"
		assert.True(t, matchComponent(pattern, fileName))
	})

	t.Run("Escape symbol in pattern, doesn't match filename", func(t *testing.T) {
		fileName := "file"
		pattern := "\\pattern"
		assert.False(t, matchComponent(pattern, fileName))
	})

	t.Run("Components match, glob", func(t *testing.T) {
		filePath := "file"
		pattern := "**"
		assert.True(t, matchComponent(pattern, filePath))
	})

	t.Run("Components match", func(t *testing.T) {
		filePath := "file"
		pattern := "file"
		assert.True(t, matchComponent(pattern, filePath))
	})

	t.Run("Components don't match", func(t *testing.T) {
		filePath := "file"
		pattern := "pattern"
		assert.False(t, matchComponent(pattern, filePath))
	})

	t.Run("Component with escape symbol", func(t *testing.T) {
		filePath := "file\\ name"
		pattern := "*"
		assert.True(t, matchComponent(pattern, filePath))
	})
}

func TestMatchComponents(t *testing.T) {
	t.Parallel()

	t.Run("Empty filepath and pattern", func(t *testing.T) {
		file := []string{}
		pattern := []string{}
		assert.True(t, matchComponents(pattern, file))
	})

	t.Run("Empty filepath, non-empty pattern", func(t *testing.T) {
		file := []string{}
		pattern := []string{"pattern"}
		assert.False(t, matchComponents(pattern, file))
	})

	t.Run("Non-empty filepath, empty pattern", func(t *testing.T) {
		file := []string{"path", "to", "file"}
		pattern := []string{}
		assert.False(t, matchComponents(pattern, file))
	})

	t.Run("Components have matching element", func(t *testing.T) {
		file := []string{
			"path",
			"to",
			"src",
			"file",
		}
		pattern := []string{"**", "src", "*"}
		assert.True(t, matchComponents(pattern, file))
	})

	t.Run("Components have matching elements", func(t *testing.T) {
		file := []string{
			"path",
			"to",
			"src",
			"file",
		}
		pattern := []string{"**", "src", "file"}
		assert.True(t, matchComponents(pattern, file))
	})

	t.Run("Components don't have matching elements", func(t *testing.T) {
		file := []string{
			"path",
			"to",
			"file.txt",
		}
		pattern := []string{"**", "src", "*"}
		assert.False(t, matchComponents(pattern, file))
	})

	t.Run("Match with single character placeholders", func(t *testing.T) {
		file := []string{
			"path",
			"to",
			"src",
			"file.go",
		}
		pattern := []string{"*", "??", "src", "file.??"}
		assert.True(t, matchComponents(pattern, file))
	})
}

func TestMatch(t *testing.T) {
	t.Parallel()

	t.Run("Exact match", func(t *testing.T) {
		fileName := "path/to/src/file"
		pattern := "path/to/src/file"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("Pattern has ** and matches filename", func(t *testing.T) {
		fileName := "path/to/src/file"
		pattern := "**/src/file"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("Invalid pattern", func(t *testing.T) {
		fileName := "path/to/src/file"
		pattern := "**src/file"
		_, err := match(pattern, fileName)
		assert.Error(t, err)
	})

	t.Run("Different paths to file with the same name", func(t *testing.T) {
		fileName := "path/to/src/file"
		pattern := "path/**/file"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("Pattern ends with / at filename end", func(t *testing.T) {
		fileName := "path/to/src/file.go"
		pattern := "**/file.go/"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("'*' matches only within a single component", func(t *testing.T) {
		fileName := "path/to/src/folder/some/files"
		pattern := "**/src/*"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.False(t, matches)
	})

	t.Run("'**' matches zero or more components in the complete file name", func(t *testing.T) {
		fileName := "path/to/src/folder/some/files"
		pattern := "**/src/**"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("Path doesn't match pattern", func(t *testing.T) {
		fileName := "path/to/file"
		pattern := "**/src/*"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.False(t, matches)
	})

	t.Run("Pattern contains **/**", func(t *testing.T) {
		fileName := "path/to/src/file"
		pattern := "path/**/**/file"
		matches, err := match(pattern, fileName)
		assert.NoError(t, err)
		assert.True(t, matches)
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
