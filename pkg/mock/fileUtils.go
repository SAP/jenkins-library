// +build !release

package mock

import (
	"io"
	"io/ioutil"
	"os"
)

type FilesMock struct {
	Files []string
}


func (f FilesMock) FileExists(filename string) (bool, error) {

	if f.Files

	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

func (f FilesMock) Copy(src, dst string) (int64, error) {

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

func (f FilesMock) FileRead(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (f FilesMock) FileWrite(path string, content []byte, perm os.FileMode) error {
	return ioutil.WriteFile(path, content, perm)
}

func (f FilesMock) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
