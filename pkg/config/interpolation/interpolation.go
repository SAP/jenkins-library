package interpolation

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	maxLookupDepth = 10
)

var (
	lookupRegex   *regexp.Regexp = regexp.MustCompile(`\$\((?P<property>[a-zA-Z0-9\.]*)\)`)
	captureGroups                = setupCaptureGroups(lookupRegex.SubexpNames())
)

// ResolveMap interpolates every string value of a map and tries to lookup references to other properties of that map
func ResolveMap(config map[string]interface{}) error {
	for key, value := range config {
		if str, ok := value.(string); ok {
			resolvedStr, err := ResolveString(str, config)
			if err != nil {
				return err
			}
			config[key] = resolvedStr
		}
	}
	return nil
}

func resolveString(str string, lookupMap map[string]interface{}, n int) (string, error) {
	matches := lookupRegex.FindAllStringSubmatch(str, -1)
	if len(matches) == 0 {
		return str, nil
	}
	if n == maxLookupDepth {
		return "", fmt.Errorf("Property could not be resolved with a depth of %d. '%s' is still left to resolve", n, str)
	}
	for _, match := range matches {
		property := match[captureGroups["property"]]
		if propVal, ok := lookupMap[property]; ok {
			str = strings.ReplaceAll(str, fmt.Sprintf("$(%s)", property), propVal.(string))
		} else {
			str = strings.ReplaceAll(str, fmt.Sprintf("$(%s)", property), "")
		}
	}
	return resolveString(str, lookupMap, n+1)
}

// ResolveString takes a string and replaces all references inside of it whith values from the given lookupMap.
// This is being done recursively until the maxLookupDepth is reached.
func ResolveString(str string, lookupMap map[string]interface{}) (string, error) {
	return resolveString(str, lookupMap, 0)
}

func setupCaptureGroups(captureGroupsList []string) map[string]int {
	groups := make(map[string]int, len(captureGroupsList))
	for i, captureGroupName := range captureGroupsList {
		if i == 0 {
			continue
		}
		groups[captureGroupName] = i
	}
	return groups
}
