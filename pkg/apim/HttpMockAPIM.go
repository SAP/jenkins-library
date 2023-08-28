//go:build !release
// +build !release

package apim

import (
	"bytes"
	"io"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
)

type HttpMockAPIM struct {
	Method       string                  // is set during test execution
	URL          string                  // is set before test execution
	Header       map[string][]string     // is set before test execution
	ResponseBody string                  // is set before test execution
	Options      piperhttp.ClientOptions // is set during test
	StatusCode   int                     // is set during test
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type SenderMock interface {
	SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	SetOptions(options piperhttp.ClientOptions)
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type ServciceKeyMock interface {
	GetServiceKey() string
}

func (c *HttpMockAPIM) SetOptions(options piperhttp.ClientOptions) {
	c.Options = options
}

func (c *HttpMockAPIM) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

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
		return &res, errors.New("Bad Request")
	}

	return &res, nil
}

func GetServiceKey() string {
	apiServiceKey := `{
		"oauth": {
			"url": "https://demo",
			"clientid": "demouser",
			"clientsecret": "******",
			"tokenurl": "https://demo/oauth/token"
			}
		}`
	return apiServiceKey
}
