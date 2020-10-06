package piperutils

import (
	"strings"
)

//ContainsInt checks whether the element is part of the slice
func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//ContainsString checks whether the element is part of the slice
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//ContainsStringPart checks whether the element is contained as part of one of the elements of the slice
func ContainsStringPart(s []string, part string) bool {
	for _, a := range s {
		if strings.Contains(a, part) {
			return true
		}
	}
	return false
}

// RemoveAll removes all instances of element from the slice and returns a truncated slice as well as
// a boolean to indicate whether at least one element was found and removed.
func RemoveAll(s []string, e string) ([]string, bool) {
	var r []string
	for _, a := range s {
		if a != e {
			r = append(r, a)
		}
	}
	return r, len(s) != len(r)
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

// SplitAndTrim iterates over the strings in the given slice and splits each on the provided separator.
// Each resulting sub-string is then a separate entry in the returned array.
func SplitAndTrim(in []string, separator string) (out []string) {
	if len(in) == 0 {
		return in
	}
	for _, entry := range in {
		entryParts := strings.Split(entry, separator)
		for _, part := range entryParts {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return
}
