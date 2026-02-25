package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// Downloader ...
type Downloader interface {
	SetOptions(options ClientOptions)
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

// DownloadFile downloads a file's content as GET request from the specified URL to the specified file
func (c *Client) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return c.DownloadRequest(http.MethodGet, url, filename, header, cookies)
}

// DownloadRequest ...
func (c *Client) DownloadRequest(method, url, filename string, header http.Header, cookies []*http.Cookie) error {
	response, err := c.SendRequest(method, url, nil, header, cookies)
	if err != nil {
		return fmt.Errorf("HTTP %v request to %v failed with error: %w", method, url, err)
	}
	defer response.Body.Close()
	parent := filepath.Dir(filename)
	if len(parent) > 0 {
		if err = os.MkdirAll(parent, 0775); err != nil {
			return err
		}
	}
	fileHandler, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create file %v: %w", filename, err)
	}
	defer fileHandler.Close()

	_, err = piperutils.CopyData(fileHandler, response.Body)
	if err != nil {
		return fmt.Errorf("unable to copy content from url to file %v: %w", filename, err)
	}
	return err
}

// GetRequest downloads content from a given URL and returns the response instead of writing it to file
func (c *Client) GetRequest(url string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	response, err := c.SendRequest("GET", url, nil, header, cookies)
	if err != nil {
		return &http.Response{}, fmt.Errorf("HTTP request to %v failed with error: %w", url, err)
	}
	return response, nil
}

// DownloadExecutable downloads a script or another executable and sets appropriate permissions
func DownloadExecutable(githubToken string, fileUtils piperutils.FileUtils, downloader Downloader, url string) (string, error) {
	header := http.Header{}
	if len(githubToken) > 0 {
		header = http.Header{"Authorization": []string{"Token " + githubToken}}
		header.Set("Accept", "application/vnd.github.v3.raw")
	}

	fileNameParts := strings.Split(url, "/")
	fileName := fileNameParts[len(fileNameParts)-1]
	fullFileName := filepath.Join(".pipeline", fileName)
	err := downloader.DownloadFile(url, fullFileName, header, []*http.Cookie{})
	if err != nil {
		return "", fmt.Errorf("unable to download script from %v: %w", url, err)
	}
	err = fileUtils.Chmod(fullFileName, 0555)
	if err != nil {
		return "", fmt.Errorf("unable to change script permission for %v: %w", fullFileName, err)
	}
	return fullFileName, nil
}
