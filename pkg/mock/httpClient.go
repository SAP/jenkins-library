//go:build !release

package mock

import (
	"fmt"
	"io"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

// HttpClientMock mock struct
type HttpClientMock struct {
	ClientOptions          []piperhttp.ClientOptions // set by mock
	HTTPFileUtils          *FilesMock
	FileUploads            map[string]string // set by mock
	ReturnFileUploadStatus int               // expected to be set upfront
	ReturnFileUploadError  error             // expected to be set upfront
}

// SendRequest mock
func (utils *HttpClientMock) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

// SetOptions mock
func (utils *HttpClientMock) SetOptions(options piperhttp.ClientOptions) {
	utils.ClientOptions = append(utils.ClientOptions, options)
}

// Upload mock
func (utils *HttpClientMock) Upload(data piperhttp.UploadRequestData) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

// UploadRequest mock
func (utils *HttpClientMock) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	utils.FileUploads[file] = url

	response := http.Response{
		StatusCode: utils.ReturnFileUploadStatus,
	}

	return &response, utils.ReturnFileUploadError
}

// UploadFile mock
func (utils *HttpClientMock) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	return utils.UploadRequest(http.MethodPut, url, file, fieldName, header, cookies, uploadType)
}

// DownloadFile mock
func (utils *HttpClientMock) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	if utils.HTTPFileUtils != nil {
		utils.HTTPFileUtils.AddFile(filename, []byte("some content"))
	}
	return nil
}
