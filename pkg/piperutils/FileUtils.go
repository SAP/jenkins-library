package piperutils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

// FindFiles ...
func FindFiles(root string, pattern string) ([]string, error) {

	var files []string

	r, err := regexp.Compile(pattern)
	if err != nil {
		return files, err
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !info.IsDir() && r.MatchString(info.Name()) {
			files = append(files, path)
		}

		return nil

	})

	return files, nil
}
