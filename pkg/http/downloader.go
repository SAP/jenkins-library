package http

import (
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

//Downloader ...
type Downloader interface {
	SetOptions(options ClientOptions)
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
}

// DownloadFile downloads a file's content as GET request from the specified URL to the specified file
func (c *Client) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return c.DownloadRequest(http.MethodPost, url, filename, header, cookies)
}

// DownloadRequest ...
func (c *Client) DownloadRequest(method, url, filename string, header http.Header, cookies []*http.Cookie) error {
	response, err := c.SendRequest(method, url, nil, header, cookies)
	if err != nil {
		return errors.Wrapf(err, "HTTP %v request to %v failed with error", method, url)
	}
	defer response.Body.Close()

	fileHandler, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "unable to create file %v", filename)
	}
	defer fileHandler.Close()

	_, err = io.Copy(fileHandler, response.Body)
	if err != nil {
		return errors.Wrapf(err, "unable to copy content from url to file %v", filename)
	}
	return err
}
