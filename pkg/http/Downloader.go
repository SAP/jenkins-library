package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Downloader interface {
	//SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	SetOptions(options ClientOptions)
	DownloadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error)
}

// UploadFile uploads a file's content as multipart-form POST request to the specified URL
func (c *Client) DownloadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return c.UploadRequest(http.MethodPost, url, file, fieldName, header, cookies)
}

// DownloadRequest ...
func (c *Client) downloadRequest(url, file, fieldName string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	if method != http.MethodGet {
		return nil, errors.New(fmt.Sprintf("Http method %v is not allowed. Possible values are %v", method, http.MethodGet))
	}

	// http.MethodGet

	response, err := c.SendRequest(http.MethodGet, url, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	nBytes, err := io.Copy(file, response.Body)
	return nBytes, err

	//bodyBuffer := &bytes.Buffer{}
	//bodyWriter := multipart.NewWriter(bodyBuffer)

	//	fileWriter, err := bodyWriter.CreateFormFile(fieldName, file)
	//	if err != nil {
	//		return &http.Response{}, errors.Wrapf(err, "error creating form file %v for field %v", file, fieldName)
	//	}

	defer response.Body.Close()
	fileHandle, err := os.Create(file)
	if err != nil {
		return 0, err
	}
	defer fileHandle.Close()

	nBytes, err := io.Copy(fileHandle, response.Body)

	fileHandle, err := os.Open(file)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "unable to locate file %v", file)
	}
	defer fileHandle.Close()

	_, err = io.Copy(fileWriter, fileHandle)
	if err != nil {
		return &http.Response{}, errors.Wrapf(err, "unable to copy file content of %v into request body", file)
	}
	err = bodyWriter.Close()

	request, err := c.createRequest(method, url, bodyBuffer, &header, cookies)
	if err != nil {
		c.logger.Debugf("New %v request to %v", method, url)
		return &http.Response{}, errors.Wrapf(err, "error creating %v request to %v", method, url)
	}

	startBoundary := strings.Index(bodyWriter.FormDataContentType(), "=") + 1
	boundary := bodyWriter.FormDataContentType()[startBoundary:]

	request.Header.Add("Content-Type", "multipart/form-data; boundary=\""+boundary+"\"")
	request.Header.Add("Connection", "Keep-Alive")

	response, err := httpClient.Do(request)
	if err != nil {
		return response, errors.Wrapf(err, "HTTP %v request to %v failed with error", method, url)
	}

	return c.handleResponse(response)
}
