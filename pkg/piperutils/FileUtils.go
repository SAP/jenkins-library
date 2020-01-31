package piperutils

import (
	"errors"
	"io"
	"os"
)

// DirectoryExists ...
func DirectoryExists(filename string) (bool, error) {
	shouldBeDir := true
	return exists(filename, shouldBeDir)
}

// FileExists ...
func FileExists(filename string) (bool, error) {
	shouldBeDir := false
	return exists(filename, shouldBeDir)
}

func exists(path string, shouldBeDir bool) (bool, error) {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir() == shouldBeDir, nil

}

// Copy ...
func Copy(src, dst string) (int64, error) {

	exists, err := FileExists(src)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, errors.New("Source file '" + src + "' does not exist")
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
