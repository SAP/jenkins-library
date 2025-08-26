//go:build !release

package mock

import (
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestSendRequest(t *testing.T) {
	t.Parallel()
	t.Run("SendRequest", func(t *testing.T) {
		utils := HttpClientMock{}
		method := "PUT"
		url := "https://localhost"
		var header http.Header
		var r io.Reader
		var cookies []*http.Cookie

		_, err := utils.SendRequest(method, url, r, header, cookies)
		assert.Error(t, err)

	})
}

func TestDownloadFile(t *testing.T) {
	t.Parallel()
	t.Run("DownloadFile", func(t *testing.T) {
		utils := HttpClientMock{
			HTTPFileUtils: &FilesMock{},
		}
		url := "https://localhost"
		filename := "testFile"
		var header http.Header
		var cookies []*http.Cookie
		err := utils.DownloadFile(url, filename, header, cookies)
		assert.NoError(t, err)
		content, err := utils.HTTPFileUtils.FileRead(filename)
		assert.NoError(t, err)
		assert.Equal(t, "some content", string(content))
	})
}

func TestSetOption(t *testing.T) {
	t.Parallel()
	t.Run("SetOption", func(t *testing.T) {
		utils := HttpClientMock{}
		options := []piperhttp.ClientOptions{
			{
				Username: "user",
				Password: "pwd",
			},
			{
				Username: "user2",
				Password: "pwd2",
			},
		}

		for _, option := range options {
			utils.SetOptions(option)
		}
		assert.Equal(t, options, utils.ClientOptions)
	})
}

func TestUpload(t *testing.T) {
	t.Parallel()
	t.Run("Upload", func(t *testing.T) {
		utils := HttpClientMock{}
		data := piperhttp.UploadRequestData{}

		_, err := utils.Upload(data)
		assert.Error(t, err)

	})
}

func TestUploadRequest(t *testing.T) {
	t.Parallel()
	t.Run("UploadRequest", func(t *testing.T) {
		utils := HttpClientMock{
			ReturnFileUploadStatus: 200,
			FileUploads: map[string]string{
				"key": "value",
			},
		}
		method := "PUT"
		url := "https://localhost"
		file := "test-7.8.9.tgz"
		fieldName := ""
		uploadType := ""
		var header http.Header
		var cookies []*http.Cookie
		returnFileUploadStatus := 200

		response, err := utils.UploadRequest(method, url, file, fieldName, header, cookies, uploadType)
		assert.NoError(t, err)
		assert.Equal(t, returnFileUploadStatus, response.StatusCode)

	})
}

func TestUploadFile(t *testing.T) {
	t.Parallel()
	t.Run("UploadFile", func(t *testing.T) {
		utils := HttpClientMock{
			ReturnFileUploadStatus: 200,
			FileUploads: map[string]string{
				"key": "value",
			},
		}
		url := "https://localhost"
		file := "test-7.8.9.tgz"
		fieldName := ""
		uploadType := ""
		var header http.Header
		var cookies []*http.Cookie
		returnFileUploadStatus := 200

		response, err := utils.UploadFile(url, file, fieldName, header, cookies, uploadType)
		assert.NoError(t, err)
		assert.Equal(t, returnFileUploadStatus, response.StatusCode)

	})
}
