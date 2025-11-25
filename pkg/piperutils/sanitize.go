package piperutils

import (
	"strings"
)

// SanitizePath removes query parameters from a URL or file path
func SanitizePath(input string) string {
	return strings.Split(input, "?")[0]
}
