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

//ContainsString check wether the element is part of the slice
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//ContainsStringPart check wether the element is contained as part of one of the elements of the slice
func ContainsStringPart(s []string, part string) bool {
	for _, a := range s {
		if strings.Contains(a, part) {
			return true
		}
	}
	return false
}
