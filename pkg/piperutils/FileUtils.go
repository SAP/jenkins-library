package piperutils

import (
	"errors"
	"io"
	"os"
)

//FileUtils ...
type FileUtils struct {
}

// FileExists ...
func (f FileUtils) FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

// FileExists ...
func FileExists(filename string) (bool, error) {
	return FileUtils{}.FileExists(filename)
}

// Copy ...
func (f FileUtils) Copy(src, dst string) (int64, error) {

	exists, err := f.FileExists(src)

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

// Copy ...
func Copy(src, dst string) (int64, error) {
	return FileUtils{}.Copy(src, dst)
}
