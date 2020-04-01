package piperutils

import (
	"strings"
)

//ContainsInt check wether the element is part of the slice
func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//Prefix adds a prefix to each element of the slice
func Prefix(in []string, prefix string) (out []string) {
	for _, element := range in {
		out = append(out, prefix+element)
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
