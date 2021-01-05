package sonar

import (
	"encoding/json"
	"net/http"

	sonargo "github.com/magicsong/sonargo/sonar"
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

// Search Search for issues.<br>At most one of the following parameters can be provided at the same time: componentKeys, componentUuids, components, componentRootUuids, componentRoots.<br>Requires the 'Browse' permission on the specified project(s).
func (s *Client) SearchIssues(opt *sonargo.IssuesSearchOption) (v *sonargo.IssuesSearchObject, resp *http.Response, err error) {
	sonarClient, err := sonargo.NewClient(s.Host, s.Username, s.Password)

	err = sonarClient.Issues.ValidateSearchOpt(opt)
	if err != nil {
		return
	}

	req, err := sonarClient.NewRequest("GET", "issues/search", opt)
	if err != nil {
		return
	}

	resp, err = s.HTTPClient.Send(req)
	if err != nil {
		return
	}

	err = sonargo.CheckResponse(resp)
	if err != nil {
		return
	}

	v = new(sonargo.IssuesSearchObject)
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(v)
	if err != nil {
		return nil, resp, err
	}
	return
}
