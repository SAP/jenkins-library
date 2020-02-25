package piperutils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// MkdirAll ...
func MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Copy ...
func Copy(src, dst string, createMissingDirectories bool) (int64, error) {

	exists, err := FileExists(src)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, errors.New("Source file '" + src + "' does not exist")
	}
	parent := filepath.Dir(dst)

	parentFolderExists, err := FileExists(parent)

	if err != nil {
		return 0, err
	}

	if !parentFolderExists {

		if !createMissingDirectories {
			return 0, fmt.Errorf("Parent folder for file '%s' does not exist, createMissingDirectories was '%t'", dst, createMissingDirectories)
		}

		// 775 will not fit always but is a reasonable default. As long as nobody complains ...
		if err = MkdirAll(parent, 0775); err != nil {
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
