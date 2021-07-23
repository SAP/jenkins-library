package blackduck

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type httpMockClient struct {
	responseBodyForURL map[string]string
	errorMessageForURL map[string]string
	header             map[string]http.Header
}

func (c *httpMockClient) SetOptions(opts piperhttp.ClientOptions) {}
func (c *httpMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.header[url] = header
	response := http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}

	if c.errorMessageForURL[url] != "" {
		response.StatusCode = 400
		return &response, fmt.Errorf(c.errorMessageForURL[url])
	}

	if c.responseBodyForURL[url] != "" {
		response.Body = ioutil.NopCloser(bytes.NewReader([]byte(c.responseBodyForURL[url])))
		return &response, nil
	}

	return &response, nil
}

const (
	authContent = `{
		"bearerToken":"bearerTestToken",
		"expiresInMilliseconds":7199997
	}`
	projectContent = `{
		"totalCount": 1,
		"items": [
			{
				"name": "SHC-PiperTest",
				"_meta": {
					"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf",
					"links": [
						{
							"rel": "versions",
							"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions"
						}
					]
				}
			}
		]
	}`
)

func TestGetProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":             authContent,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": projectContent,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		project, err := bdClient.GetProject("SHC-PiperTest")
		assert.NoError(t, err)
		assert.Equal(t, "SHC-PiperTest", project.Name)
		assert.Equal(t, "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions", project.Metadata.Links[0].Href)
		headerExpected := http.Header{"Authorization": []string{"Bearer bearerTestToken"}, "Accept": {"application/vnd.blackducksoftware.project-detail-4+json"}}
		assert.Equal(t, headerExpected, myTestClient.header["https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest"])
	})

	t.Run("failure - not found", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
			},
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": "not found",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProject("SHC-PiperTest")
		assert.Contains(t, fmt.Sprint(err), "failed to get project 'SHC-PiperTest'")
	})

	t.Run("failure - 0 results", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": `{
					"totalCount": 0,
					"items": []
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProject("SHC-PiperTest")
		assert.Contains(t, fmt.Sprint(err), "project 'SHC-PiperTest' not found")
	})

	t.Run("failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":             authContent,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProject("SHC-PiperTest")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve details for project 'SHC-PiperTest'")
	})
}

func TestGetProjectVersion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":             authContent,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions": `{
					"totalCount": 1,
					"items": [
						{
							"versionName": "1.0",
							"_meta": {
								"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36",
								"links": []
							}
						}
					]
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		projectVersion, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.Equal(t, "1.0", projectVersion.Name)
		assert.Equal(t, "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36", projectVersion.Metadata.Href)
		headerExpected := http.Header{"Authorization": []string{"Bearer bearerTestToken"}, "Accept": {"application/vnd.blackducksoftware.project-detail-4+json"}}
		assert.Equal(t, headerExpected, myTestClient.header["https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions"])
	})

	t.Run("failure - project not found", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
			},
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": "not found",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to get project 'SHC-PiperTest'")
	})

	t.Run("failure - version not found", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":             authContent,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": projectContent,
			},
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions": "not found",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to get project version 'SHC-PiperTest:1.0'")
	})

	t.Run("failure - 0 results", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":             authContent,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions": `{
					"totalCount": 0,
					"items": []
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "project version 'SHC-PiperTest:1.0' not found")
	})

	t.Run("failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                    authContent,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest":                        projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve details for project version 'SHC-PiperTest:1.0'")
	})
}

func TestSendRequestIntegration(t *testing.T) {
	token := "ZjNiZTYxOWYtYjIyYi00YzZkLTk3YTAtYzZjYjU0ZTkxZmY0OmEyMjEzZGQzLTdlMWQtNDkyNy1hZTkzLThjZTQyNjdkYjBhNA=="
	bdClient := NewClient(token, "https://sap.blackducksoftware.com", &piperhttp.Client{})

	res, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")

	assert.NoError(t, err)
	t.Log(res)
}

func TestAuthenticate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		err := bdClient.authenticate()
		assert.NoError(t, err)
		headerExpected := http.Header{"Authorization": {"token myTestToken"}, "Accept": {"application/vnd.blackducksoftware.user-4+json"}}
		assert.Equal(t, headerExpected, myTestClient.header["https://my.blackduck.system/api/tokens/authenticate"])
	})

	t.Run("authentication failure", func(t *testing.T) {
		myTestClient := httpMockClient{
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": "not authorized",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		err := bdClient.authenticate()
		assert.EqualError(t, err, "authentication to BlackDuck API failed: request to BlackDuck API failed: not authorized")
	})

	t.Run("parse failure", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		err := bdClient.authenticate()
		assert.Contains(t, fmt.Sprint(err), "failed to parse BlackDuck response")
	})
}

func TestSendRequest(t *testing.T) {
	myTestClient := httpMockClient{
		responseBodyForURL: map[string]string{
			"https://my.blackduck.system/api/endpoint":        "testContent",
			"https://my.blackduck.system/api/endpoint?q=test": "testContentQuery",
		},
		header: map[string]http.Header{},
	}
	bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)

	t.Run("simple", func(t *testing.T) {
		responseBody, err := bdClient.sendRequest(http.MethodGet, "api/endpoint", map[string]string{}, nil, http.Header{})
		assert.NoError(t, err)
		assert.Equal(t, "testContent", string(responseBody))
	})

	t.Run("with query params and bearer", func(t *testing.T) {
		bdClient.BearerToken = "testBearer"
		responseBody, err := bdClient.sendRequest(http.MethodGet, "api/endpoint", map[string]string{"q": "test"}, nil, http.Header{})
		assert.NoError(t, err)
		assert.Equal(t, "testContentQuery", string(responseBody))
		assert.Equal(t, http.Header{"Authorization": {"Bearer testBearer"}}, myTestClient.header["https://my.blackduck.system/api/endpoint?q=test"])
	})
}

func TestApiURL(t *testing.T) {
	tt := []struct {
		description string
		apiEndpoint string
		client      Client
		expected    string
	}{
		{
			description: "trailing / in path",
			apiEndpoint: "/my/path/",
			client:      Client{serverURL: "https://my.test.server"},
			expected:    "https://my.test.server/my/path",
		},
		{
			description: "trailing / in server",
			apiEndpoint: "/my/path",
			client:      Client{serverURL: "https://my.test.server/"},
			expected:    "https://my.test.server/my/path",
		},
	}

	for _, test := range tt {
		res, err := test.client.apiURL(test.apiEndpoint)
		assert.NoError(t, err)
		assert.Equalf(t, test.expected, res.String(), test.description)
	}
}

func TestAuthenticationValid(t *testing.T) {
	tt := []struct {
		description string
		now         time.Time
		client      Client
		expected    bool
	}{
		{
			description: "login still valid",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 59, 0, 0, time.UTC)},
			expected:    true,
		},
		{
			description: "login still valid - edge",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 58, 1, 0, time.UTC)},
			expected:    true,
		},
		{
			description: "login expired",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 57, 0, 0, time.UTC)},
			expected:    false,
		},
		{
			description: "login expired - edge",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 58, 0, 0, time.UTC)},
			expected:    false,
		},
	}

	for _, test := range tt {
		assert.Equalf(t, test.expected, test.client.authenticationValid(test.now), test.description)
	}
}

func TestUrlPath(t *testing.T) {
	assert.Equal(t, "/this/is/the/path", urlPath("https://the.server.domain:8080/this/is/the/path"))
}
