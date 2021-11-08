package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

//Downloader ...
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
		return errors.Wrapf(err, "HTTP %v request to %v failed with error", method, url)
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
		return errors.Wrapf(err, "unable to create file %v", filename)
	}
	defer fileHandler.Close()

	_, err = piperutils.CopyData(fileHandler, response.Body)
	if err != nil {
		return errors.Wrapf(err, "unable to copy content from url to file %v", filename)
	}
	return err
}

// GetRequest downloads content from a given URL and returns the response instead of writing it to file
func (c *Client) GetRequest(url string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	// Downloads content from URL and returns the data instead of writing it to file
	response, err := c.SendRequest("GET", url, nil, header, cookies)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "HTTP request to %v failed with error", url)
	}
	return response, nil
}
