package piperutils

import (
	"errors"
	"io"
	"os"
)

// FileExists ...
func FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

// Copy ...
func Copy(src, dst string) (int64, error) {

	if exists, err := FileExists(src); exists && err == nil {
		errors.New("Source file '" + src + "' does not exist")
	} else {
		if err != nil {
			return 0, err
		}
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
