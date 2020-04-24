package piperutils

import (
	"strings"
)

//ContainsInt check whether the element is part of the slice
func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//ContainsString check whether the element is part of the slice
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//Prefix adds a prefix to each element of the slice
func Prefix(in []string, prefix string) []string {
	return _prefix(in, prefix, true)
}

//PrefixIfNeeded adds a prefix to each element of the slice if not already prefixed
func PrefixIfNeeded(in []string, prefix string) []string {
	return _prefix(in, prefix, false)
}

func _prefix(in []string, prefix string, always bool) (out []string) {
	for _, element := range in {
		if always || !strings.HasPrefix(element, prefix) {
			element = prefix + element
		}
		out = append(out, element)
	}
	return
}

//Trim removes dangling whitespaces from each element of the slice, empty elements are dropped
func Trim(in []string) (out []string) {
	for _, element := range in {
		if trimmed := strings.TrimSpace(element); len(trimmed) > 0 {
			out = append(out, trimmed)
		}
	}
	return
}

// SplitTrimAndDeDup iterates over the strings in the given slice and splits each on the provided separator.
// Each resulting sub-string is then a separate entry in the returned array. Duplicate and empty entries are eliminated.
func SplitTrimAndDeDup(in []string, separator string) (out []string) {
	if len(in) == 0 {
		return in
	}
	for _, entry := range in {
		entryParts := strings.Split(entry, separator)
		for _, part := range entryParts {
			part = strings.TrimSpace(part)
			if part != "" && !ContainsString(out, part) {
				out = append(out, part)
			}
		}
	}
	return
}
