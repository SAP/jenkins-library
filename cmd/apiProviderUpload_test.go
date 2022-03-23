package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRunApiProviderUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("test.json", []byte("Test content"))
		clientOptions := piperhttp.ClientOptions{}
		clientOptions.Token = fmt.Sprintf("Bearer %s", "Demo")
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := apiProviderUploadOptions{
			APIServiceKey: apiServiceKey,
			FilePath:      "test.json",
		}

		httpClient := httpMockSimple{StatusCode: 201, Options: clientOptions, ResponseBody: ``}

		// test
		err := createApiProvider(&config, nil, &httpClient, "", filesMock.FileRead)

		// assert
		if assert.NoError(t, err) {
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "/apiportal/api/1.0/Management.svc/APIProviders", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Failure path", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("test.json", []byte("Test content"))
		clientOptions := piperhttp.ClientOptions{}
		clientOptions.Token = fmt.Sprintf("Bearer %s", "Demo")
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := apiProviderUploadOptions{
			APIServiceKey: apiServiceKey,
			FilePath:      "test.json",
		}

		httpClient := httpMockSimple{StatusCode: 400, Options: clientOptions, ResponseBody: ``}

		// test
		err := createApiProvider(&config, nil, &httpClient, "", filesMock.FileRead)

		// assert
		assert.EqualError(t, err, "HTTP \"POST\" request to \"/apiportal/api/1.0/Management.svc/APIProviders\" failed with error: a http error occurred")
	})

}

type httpMockSimple struct {
	Method       string                  // is set during test execution
	URL          string                  // is set before test execution
	ResponseBody string                  // is set before test execution
	Options      piperhttp.ClientOptions // is set during test
	StatusCode   int                     // is set during test
}

func (c *httpMockSimple) SetOptions(options piperhttp.ClientOptions) {
	c.Options = options
}

func (c *httpMockSimple) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	c.Method = method
	c.URL = url

	if r != nil {
		_, err := ioutil.ReadAll(r)

		if err != nil {
			return nil, err
		}
	}

	if c.Options.Token == "" {
		c.ResponseBody = "{\r\n\t\t\t\"access_token\": \"demotoken\",\r\n\t\t\t\"token_type\": \"Bearer\",\r\n\t\t\t\"expires_in\": 3600,\r\n\t\t\t\"scope\": \"\"\r\n\t\t}"
		c.StatusCode = 200
		res := http.Response{
			StatusCode: c.StatusCode,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(c.ResponseBody))),
		}
		return &res, nil
	}

	res := http.Response{
		StatusCode: c.StatusCode,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(c.ResponseBody))),
	}

	if c.StatusCode >= 400 {
		return &res, errors.New("a http error occurred")
	}

	return &res, nil
}
