package sonar

import (
	"encoding/json"
	"net/http"

	sonargo "github.com/magicsong/sonargo/sonar"
	"github.com/prometheus/common/log"
)

type Client struct {
	Username   string
	Password   string
	Host       string
	HTTPClient Sender
	// Certificates [][]byte
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	Send(*http.Request) (*http.Response, error)
	// SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)
	// SetOptions(options piperHttp.ClientOptions)
}

func NewBasicAuthClient(username, password, host string, client Sender) *Client {
	return &Client{
		Username:   username,
		Password:   password,
		Host:       host,
		HTTPClient: client,
	}
}

// SearchIssues Search for issues.<br>At most one of the following parameters can be provided at the same time: componentKeys, componentUuids, components, componentRootUuids, componentRoots.<br>Requires the 'Browse' permission on the specified project(s).
func (s *Client) SearchIssues(options *sonargo.IssuesSearchOption) (result *sonargo.IssuesSearchObject, response *http.Response, err error) {
	sonarClient, err := sonargo.NewClient(s.Host, s.Username, s.Password)
	// reuse parameter validation from sonargo
	err = sonarClient.Issues.ValidateSearchOpt(options)
	if err != nil {
		return
	}
	// reuse request creation from sonargo
	req, err := sonarClient.NewRequest("GET", "issues/search", options)
	if err != nil {
		return
	}
	// request created by sonarGO uses .Opaque without the host parameter leading to a request against https://api/issues/search
	// https://github.com/magicsong/sonargo/blob/103eda7abc20bd192a064b6eb94ba26329e339f1/sonar/sonarqube.go#L55
	req.URL.Opaque = ""
	req.URL.Path = sonarClient.BaseURL().Path + "issues/search"
	log.Warnf("REQUEST: %v", req)
	// use custom HTTP client to send request
	response, err = s.HTTPClient.Send(req)
	if err != nil {
		return
	}
	// reuse response verrification from sonargo
	err = sonargo.CheckResponse(response)
	if err != nil {
		return
	}
	// decode JSON response
	result = new(sonargo.IssuesSearchObject)
	err = s.decode(response, result)
	if err != nil {
		return nil, response, err
	}
	return
}

func (s *Client) decode(resp *http.Response, v interface{}) (err error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
