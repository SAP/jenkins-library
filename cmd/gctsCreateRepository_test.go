//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestGctsCreateRepositorySuccess(t *testing.T) {

	config := gctsCreateRepositoryOptions{
		Host:                "http://testHost.com:50000",
		Client:              "000",
		Repository:          "testRepo",
		Username:            "testUser",
		Password:            "testPassword",
		RemoteRepositoryURL: "https://github.com/org/testRepo",
		Role:                "SOURCE",
		VSID:                "TST",
	}

	t.Run("creating repository on ABAP system successful", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `{
			"repository": {
				"rid": "my-repository",
				"name": "Example repository",
				"role": "SOURCE",
				"type": "GIT",
				"vsid": "GI7",
				"status": "READY",
				"branch": "master",
				"url": "https://github.com/git/git",
				"version": "1.0.1",
				"objects": 1337,
				"currentCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
				"connection": "ssl",
				"config": [
					{
						"key": "CLIENT_VCS_URI",
						"value": "git@github.com/example.git"
					}
				]
			},
			"log": [
				{
					"time": 20180606130524,
					"user": "JENKINS",
					"section": "REPOSITORY_FACTORY",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "Start action CREATE_REPOSITORY review",
					"code": "GCTS.API.410"
				}
			]
		}`}

		err := createRepository(&config, nil, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})

			t.Run("check user", func(t *testing.T) {
				assert.Equal(t, "testUser", httpClient.Options.Username)
			})

			t.Run("check password", func(t *testing.T) {
				assert.Equal(t, "testPassword", httpClient.Options.Password)
			})

		}

	})

	t.Run("repository already exists on ABAP system", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `{
			"exception": "Repository already exists"
		}`}

		err := createRepository(&config, nil, nil, &httpClient)

		assert.NoError(t, err)
	})
}
func TestGctsCreateRepositoryFailure(t *testing.T) {

	config := gctsCreateRepositoryOptions{
		Host:                "http://testHost.com:50000",
		Client:              "000",
		Repository:          "testRepo",
		Username:            "testUser",
		Password:            "testPassword",
		RemoteRepositoryURL: "https://github.com/org/testRepo",
		Role:                "SOURCE",
		VSID:                "TST",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `{
			"log": [
				{
					"time": 20180606130524,
					"user": "JENKINS",
					"section": "REPOSITORY_FACTORY",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "Start action CREATE_REPOSITORY review",
					"code": "GCTS.API.410"
				}
			],
			"errorLog": [
				{
					"time": 20180606130524,
					"user": "JENKINS",
					"section": "REPOSITORY_FACTORY",
					"action": "CREATE_REPOSITORY",
					"severity": "INFO",
					"message": "Start action CREATE_REPOSITORY review",
					"code": "GCTS.API.410"
				}
			],
			"exception": {
				"message": "repository_not_found",
				"description": "Repository not found",
				"code": 404
			}
		}`}

		err := createRepository(&config, nil, nil, &httpClient)

		assert.EqualError(t, err, "creating repository on the ABAP system http://testHost.com:50000 failed: a http error occurred")
	})
}

type httpMockGcts struct {
	Method       string                  // is set during test execution
	URL          string                  // is set before test execution
	Header       map[string][]string     // is set before test execution
	ResponseBody string                  // is set before test execution
	Options      piperhttp.ClientOptions // is set during test
	StatusCode   int                     // is set during test
}

func (c *httpMockGcts) SetOptions(options piperhttp.ClientOptions) {
	c.Options = options
}

func (c *httpMockGcts) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	c.Method = method
	c.URL = url

	if r != nil {
		_, err := io.ReadAll(r)

		if err != nil {
			return nil, err
		}
	}

	res := http.Response{
		StatusCode: c.StatusCode,
		Header:     c.Header,
		Body:       io.NopCloser(bytes.NewReader([]byte(c.ResponseBody))),
	}

	if c.StatusCode >= 400 {
		return &res, errors.New("a http error occurred")
	}

	return &res, nil
}
