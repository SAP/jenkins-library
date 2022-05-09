package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/apim"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRunApiProviderUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		file, tmpErr := ioutil.TempFile("", "test.json")
		if tmpErr != nil {
			t.FailNow()
		}
		defer os.RemoveAll(file.Name()) // clean up
		filesMock := mock.FilesMock{}
		filesMock.AddFile(file.Name(), []byte("Test content"))
		config := getDefaultOptionsForApiProvider()
		config.FilePath = file.Name()
		httpClientMock := httpMockAPIM{StatusCode: 201, ResponseBody: ``}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: &httpClientMock}
		// test
		err := createApiProvider(&config, apim, filesMock.FileRead)

		// assert
		if assert.NoError(t, err) {
			t.Run("check file", func(t *testing.T) {
				fileUtils = &piperutils.Files{}
				fExists, err := fileUtils.FileExists(file.Name())
				assert.NoError(t, err)
				assert.Equal(t, fExists, true)
			})

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "/apiportal/api/1.0/Management.svc/APIProviders", httpClientMock.URL)
			})
			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClientMock.Method)
			})
		}
	})

	t.Run("Failure Path", func(t *testing.T) {
		file, tmpErr := ioutil.TempFile("", "test.json")
		if tmpErr != nil {
			t.FailNow()
		}
		defer os.RemoveAll(file.Name()) // clean up
		filesMock := mock.FilesMock{}
		filesMock.AddFile(file.Name(), []byte("Test content"))
		config := getDefaultOptionsForApiProvider()
		config.FilePath = file.Name()
		httpClientMock := httpMockAPIM{StatusCode: 400, ResponseBody: ``}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: &httpClientMock}
		// test
		err := createApiProvider(&config, apim, filesMock.FileRead)
		// assert
		assert.Error(t, err)
	})

}

func getDefaultOptionsForApiProvider() apiProviderUploadOptions {
	return apiProviderUploadOptions{
		APIServiceKey: `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`,
		FilePath: "test.json",
	}
}

type httpMockAPIM struct {
	Method       string                  // is set during test execution
	URL          string                  // is set before test execution
	Header       map[string][]string     // is set before test execution
	ResponseBody string                  // is set before test execution
	Options      piperhttp.ClientOptions // is set during test
	StatusCode   int                     // is set during test
}

func (c *httpMockAPIM) SetOptions(options piperhttp.ClientOptions) {
	c.Options = options
}

func (c *httpMockAPIM) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	c.Method = method
	c.URL = url

	if r != nil {
		_, err := ioutil.ReadAll(r)

		if err != nil {
			return nil, err
		}
	}

	res := http.Response{
		StatusCode: c.StatusCode,
		Header:     c.Header,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(c.ResponseBody))),
	}

	if c.StatusCode >= 400 {
		return &res, errors.New("a http error occurred")
	}

	return &res, nil
}
