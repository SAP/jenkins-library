package piperutils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
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

// Download ...
func Download(url, filename string) (int64, error) {
	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	// non-2xx codes do not create an error
	if response.StatusCode < 200 && response.StatusCode >= 300 {
		return 0, fmt.Errorf("Request failed with status code %v", response.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	nBytes, err := io.Copy(file, response.Body)
	return nBytes, err
}
