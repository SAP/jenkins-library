package helper

import (
	"fmt"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occured: %v\n", err)
		os.Exit(1)
	}
}

func contains(v []string, s string) bool {
	for _, i := range v {
		if i == s {
			return true
		}
	}
	return false
}

func ifThenElse(condition bool, positive string, negative string) string {
	if condition {
		return positive
	}
	return negative
}
