package piperutils

import (
	"os"
)

// FileExists ...
func FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)

	if err != nil {
		return false, err
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return !info.IsDir(), nil
}
