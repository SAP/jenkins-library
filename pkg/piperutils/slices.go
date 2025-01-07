package piperutils

import (
	"reflect"
	"strings"
)

// ContainsStringPart checks whether the element is contained as part of one of the elements of the slice
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
func RemoveAll(arr []string, item string) ([]string, bool) {
	res := make([]string, 0, len(arr))
	for _, a := range arr {
		if a != item {
			res = append(res, a)
		}
	}
	return res, len(arr) != len(res)
}

// Prefix adds a prefix to each element of the slice
func Prefix(in []string, prefix string) []string {
	return _prefix(in, prefix, true)
}

// PrefixIfNeeded adds a prefix to each element of the slice if not already prefixed
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

// Trim removes dangling whitespaces from each element of the slice, empty elements are dropped
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

// UniqueStrings removes duplicates from values
func UniqueStrings(values []string) []string {

	u := map[string]bool{}
	for _, e := range values {
		u[e] = true
	}
	keys := make([]string, len(u))
	i := 0
	for k := range u {
		keys[i] = k
		i++
	}
	return keys
}

// CopyAtoB copies the contents of a into slice b given that they are of equal size and compatible type
func CopyAtoB(a, b interface{}) {
	src := reflect.ValueOf(a)
	tgt := reflect.ValueOf(b)
	if src.Kind() != reflect.Slice || tgt.Kind() != reflect.Slice {
		panic("CopyAtoB() given a non-slice type")
	}

	if src.Len() != tgt.Len() {
		panic("CopyAtoB() given non equal sized slices")
	}

	// Keep the distinction between nil and empty slice input
	if src.IsNil() {
		return
	}

	for i := 0; i < src.Len(); i++ {
		tgt.Index(i).Set(src.Index(i))
	}
}
