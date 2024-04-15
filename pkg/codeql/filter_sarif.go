package codeql

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/bmatcuk/doublestar"
)

type Pattern struct {
	sign        bool
	filePattern string
	rulePattern string
}

func ParsePatterns(filterPatterns []string) ([]*Pattern, error) {
	patterns := []*Pattern{}
	for _, pattern := range filterPatterns {
		parsedPattern, err := parsePattern(pattern)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, parsedPattern)
		log.Entry().Infof("files: %s, rules: %s (include: %t)", parsedPattern.filePattern, parsedPattern.rulePattern, parsedPattern.sign)
	}
	return patterns, nil
}

func parsePattern(line string) (*Pattern, error) {
	sign, pattern := getSignAndTrimPattern(line)
	filePattern, rulePattern, err := separateFileAndRulePattern(pattern)
	if err != nil {
		return nil, err
	}
	if rulePattern == "" {
		rulePattern = "**"
	}
	log.Entry().Infof("rulePattern %s, filePattern %s", rulePattern, filePattern)
	return &Pattern{
		sign:        sign,
		filePattern: filePattern,
		rulePattern: rulePattern,
	}, nil
}

func getSignAndTrimPattern(pattern string) (bool, string) {
	sign := true

	if strings.HasPrefix(pattern, "-") {
		sign = false
		pattern = strings.TrimPrefix(pattern, "-")
	} else {
		pattern = strings.TrimPrefix(pattern, "+")
	}

	return sign, pattern
}

func separateFileAndRulePattern(pattern string) (string, string, error) {
	escChar := '\\'
	sepChar := ':'
	filePattern := ""
	rulePattern := ""
	seenSeparator := false

	for i := 0; i < len(pattern); i++ {
		c := rune(pattern[i])

		if c == sepChar {
			if seenSeparator {
				return "", "", fmt.Errorf("Invalid pattern: '%s'. Contains more than one separator!\n", pattern)
			}
			seenSeparator = true
			continue
		} else if c == escChar {
			// If we find an escape character and the current position
			// is less than total length - 1, we get the next character instead
			var nextC rune
			if i+1 < len(pattern) {
				nextC = rune(pattern[i+1])
			}
			if nextC == '+' || nextC == '-' || nextC == escChar || nextC == sepChar {
				i++
				c = nextC
			}
		}

		if seenSeparator {
			rulePattern += string(c)
		} else {
			filePattern += string(c)
		}
	}

	return filePattern, rulePattern, nil
}

func FilterSarif(sarif map[string]interface{}, patterns []*Pattern) (map[string]interface{}, error) {
	runs, ok := sarif["runs"].([]interface{})
	if !ok {
		return sarif, nil
	}

	for _, run := range runs {
		runMap, ok := run.(map[string]interface{})
		if !ok {
			continue
		}

		results, ok := runMap["results"].([]interface{})
		if !ok {
			continue
		}

		newResults := []interface{}{}
		for _, result := range results {
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				continue
			}

			locations, ok := resultMap["locations"].([]interface{})
			if !ok {
				continue
			}

			newLocations := []interface{}{}
			for _, location := range locations {
				locationMap, ok := location.(map[string]interface{})
				if !ok {
					continue
				}

				uri, ok := locationMap["physicalLocation"].(map[string]interface{})["artifactLocation"].(map[string]interface{})["uri"].(string)
				if !ok {
					continue
				}
				ruleId := resultMap["ruleId"].(string)
				include, err := matchPathAndRule(uri, ruleId, patterns)
				if err != nil {
					return nil, err
				}

				if uri != "" && include {
					log.Entry().Infof("added location to results: %s", uri)
					newLocations = append(newLocations, location)
				} else {
					log.Entry().Infof("removed %v from results", uri)
				}
			}

			if len(newLocations) > 0 {
				resultMap["locations"] = newLocations
				newResults = append(newResults, result)
			}
		}
		runMap["results"] = newResults
	}

	return sarif, nil
}

func matchPathAndRule(path string, ruleId string, patterns []*Pattern) (bool, error) {
	include := true
	for _, p := range patterns {
		matchedPath, err := doublestar.PathMatch(p.filePattern, path)
		if err != nil {
			return false, err
		}
		matchedRule, err := doublestar.Match(p.rulePattern, ruleId)
		if err != nil {
			return false, err
		}
		if matchedPath && matchedRule {
			include = p.sign
		}
	}
	return include, nil
}

func ReadSarifFile(input string) (map[string]interface{}, error) {
	var sarif map[string]interface{}
	file, err := os.Open(input)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to open sarif file: %s", err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sarif)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the JSON: %s", err)
	}
	return sarif, nil
}

func WriteSarifFile(output string, sarif map[string]interface{}) error {
	file, err := os.Create(output)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to create filtered sarif file: %s", err)
	}
	writer := json.NewEncoder(file)
	writer.SetIndent("", "    ")
	err = writer.Encode(sarif)
	if err != nil {
		return fmt.Errorf("failed to encode filtered sarif file: %s", err)
	}
	log.Entry().Infof("Successfully written filtered sarif file to %s", output)
	return nil
}
