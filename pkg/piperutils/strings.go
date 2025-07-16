package piperutils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

func Title(in string) string {
	return cases.Title(language.English, cases.NoLower).String(in)
}

func StringWithDefault(input, defaultValue string) string {
	inputCleared := strings.TrimSpace(input)
	if inputCleared == "" {
		return defaultValue
	}
	return inputCleared
}
