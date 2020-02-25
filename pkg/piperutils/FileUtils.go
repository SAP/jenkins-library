package piperutils

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
)

// FileUtils ...
type FileUtils interface {
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
}

// Files ...
type Files struct {
}

// FileExists ...
func (f Files) FileExists(filename string) (bool, error) {
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
	return Files{}.FileExists(filename)
}

// Copy ...
func (f Files) Copy(src, dst string) (int64, error) {

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
	return Files{}.Copy(src, dst)
}

//FileRead ...
func (f Files) FileRead(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// FileWrite ...
func (f Files) FileWrite(path string, content []byte, perm os.FileMode) error {
	return ioutil.WriteFile(path, content, perm)
}

// MkdirAll ...
func (f Files) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
