// +build !release

package mock

import (
	"errors"
	"os"
)

// FilesMock ...
type FilesMock struct {
	Files []string
}

// FileExists ...
func (f FilesMock) FileExists(filename string) (bool, error) {

	for _, file := range f.Files {
		if file == filename {
			return true, nil
		}

	}

	return false, nil
}

// Copy ...
func (f FilesMock) Copy(src, dst string) (int64, error) {
	return 0, errors.New("Not implemented")
}

// FileRead ...
func (f FilesMock) FileRead(path string) ([]byte, error) {
	return nil, errors.New("Not implemented")
}

// FileWrite ...
func (f FilesMock) FileWrite(path string, content []byte, perm os.FileMode) error {
	return errors.New("Not implemented")
}

// MkdirAll ...
func (f FilesMock) MkdirAll(path string, perm os.FileMode) error {
	return errors.New("Not implemented")
}
