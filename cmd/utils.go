package cmd

import (
	"os"
)

// Deprecated: Please use piperutils.Files{} instead
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
