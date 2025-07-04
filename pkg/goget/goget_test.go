//go:build unit

package goget

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestGoGetClient(t *testing.T) {
	t.Parallel()

	type expectations struct {
		registryURL  string
		errorMessage string
	}
	tests := []struct {
		name          string
		goModuleURL   string
		vcs           string
		registryURL   string
		returnStatus  int
		returnContent string
		expect        expectations
	}{
		{
			name:         "success case",
			goModuleURL:  "example.com/my/repo",
			vcs:          "git",
			registryURL:  "https://git.example.com/my/repo.git",
			returnStatus: 200,
			expect: expectations{
				registryURL: "https://git.example.com/my/repo.git",
			},
		},
		{
			name:         "error - module doesn't exist",
			goModuleURL:  "example.com/my/repo",
			vcs:          "git",
			returnStatus: 404,
			expect:       expectations{errorMessage: "module 'example.com/my/repo' doesn't exist"},
		},
		{
			name:         "error - unexpected status code",
			goModuleURL:  "example.com/my/repo",
			vcs:          "git",
			returnStatus: 401,
			expect:       expectations{errorMessage: "received unexpected response status code: 401"},
		},
		{
			name:          "error - endpoint doesn't implement the go-import protocol",
			returnStatus:  200,
			returnContent: "<!DOCTYPE html>\n<html lang=\"en\"><head></head></html>",
			expect:        expectations{errorMessage: "couldn't find go-import statement"},
		},
		{
			name:         "error - unsupported vcs",
			returnStatus: 200,
			goModuleURL:  "example.com/my/repo",
			vcs:          "svn",
			registryURL:  "https://svn.example.com/my/repo/trunk",
			expect:       expectations{errorMessage: "unsupported module: 'example.com/my/repo'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnContent := tt.returnContent

			if returnContent == "" {
				returnContent = fmt.Sprintf("<!DOCTYPE html>\n<html lang=\"en\"><head><meta name=\"go-import\" content=\"%s %s %s\"></head></html>", tt.goModuleURL, tt.vcs, tt.registryURL)
			}

			goget := ClientImpl{
				HTTPClient: &httpMock{StatusCode: tt.returnStatus, ResponseBody: returnContent},
			}

			repo, err := goget.GetRepositoryURL(tt.goModuleURL)

			if tt.expect.errorMessage != "" {
				assert.EqualError(t, err, tt.expect.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect.registryURL, repo)
			}
		})
	}
}

type httpMock struct {
	Method       string                  // is set during test execution
	URL          string                  // is set before test execution
	ResponseBody string                  // is set before test execution
	Options      piperhttp.ClientOptions // is set during test
	StatusCode   int                     // is set during test
	Body         readCloserMock          // is set during test
	Header       http.Header             // is set during test
}

func (c *httpMock) SetOptions(options piperhttp.ClientOptions) {
	c.Options = options
}

func (c *httpMock) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.Method = method
	c.URL = url
	c.Header = header

	if r != nil {
		_, err := io.ReadAll(r)

		if err != nil {
			return nil, err
		}
	}

	c.Body = readCloserMock{Content: c.ResponseBody}
	res := http.Response{StatusCode: c.StatusCode, Body: &c.Body}

	return &res, nil
}

type readCloserMock struct {
	Content string
	Closed  bool
}

func (rc readCloserMock) Read(b []byte) (n int, err error) {

	if len(b) < len(rc.Content) {
		// in real life we would fill the buffer according to buffer size ...
		return 0, fmt.Errorf("Buffer size (%d) not sufficient, need: %d", len(b), len(rc.Content))
	}
	copy(b, rc.Content)
	return len(rc.Content), io.EOF
}

func (rc *readCloserMock) Close() error {
	rc.Closed = true
	return nil
}
