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
func Prefix(slice []string, prefix string) []string {
	for idx, element := range slice {
		element = strings.TrimSpace(element)
		if !strings.HasPrefix(element, prefix) {
			element = prefix + element
		}
		slice[idx] = element
	}
	return slice
}
