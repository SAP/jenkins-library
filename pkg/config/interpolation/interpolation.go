package interpolation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	maxLookupDepth = 10
)

var (
	lookupRegex   *regexp.Regexp = regexp.MustCompile(`\$\((?P<property>[a-zA-Z0-9\.]*)\)`)
	captureGroups                = setupCaptureGroups(lookupRegex.SubexpNames())
)

// ResolveMap interpolates every string value of a map and tries to lookup references to other properties of that map
func ResolveMap(config map[string]interface{}) bool {
	for key, value := range config {
		if str, ok := value.(string); ok {
			resolvedStr, ok := ResolveString(str, config)
			if !ok {
				return false
			}
			config[key] = resolvedStr
		}
	}
	return true
}

func resolveString(str string, lookupMap map[string]interface{}, n int) (string, bool) {
	matches := lookupRegex.FindAllStringSubmatch(str, -1)
	if len(matches) == 0 {
		return str, true
	}
	if n == maxLookupDepth {
		log.Entry().Errorf("Property could not be resolved with a depth of %d. '%s' is still left to resolve", n, str)
		return "", false
	}
	for _, match := range matches {
		property := match[captureGroups["property"]]
		if propVal, ok := lookupMap[property]; ok {
			str = strings.ReplaceAll(str, fmt.Sprintf("$(%s)", property), propVal.(string))
		} else {
			// value not found
			return "", false
		}
	}
	return resolveString(str, lookupMap, n+1)
}

// ResolveString takes a string and replaces all references inside of it with values from the given lookupMap.
// This is being done recursively until the maxLookupDepth is reached.
func ResolveString(str string, lookupMap map[string]interface{}) (string, bool) {
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
