package sonar

import (
	"encoding/json"
	"net/http"

	sonargo "github.com/magicsong/sonargo/sonar"
)

// Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

type Requester struct {
	Host      string
	BasicAuth *BasicAuth
	Client    Sender
	// Certificates [][]byte
	// CACert    []byte
	// SslVerify bool
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	Send(*http.Request) (*http.Response, error)
}

func NewBasicAuthClient(username, password, host string, client Sender) *Requester {
	return &Requester{
		Host:      host,
		BasicAuth: &BasicAuth{Username: username, Password: password},
		Client:    client,
	}
}

// SearchIssues Search for issues.<br>At most one of the following parameters can be provided at the same time: componentKeys, componentUuids, components, componentRootUuids, componentRoots.<br>Requires the 'Browse' permission on the specified project(s).
func (s *Requester) SearchIssues(options *IssuesSearchOption) (result *sonargo.IssuesSearchObject, response *http.Response, err error) {
	sonarClient, err := sonargo.NewClient(s.Host, s.BasicAuth.Username, s.BasicAuth.Password)
	// reuse request creation from sonargo
	req, err := sonarClient.NewRequest("GET", "issues/search", options)
	if err != nil {
		return
	}
	// request created by sonarGO uses .Opaque without the host parameter leading to a request against https://api/issues/search
	// https://github.com/magicsong/sonargo/blob/103eda7abc20bd192a064b6eb94ba26329e339f1/sonar/sonarqube.go#L55
	req.URL.Opaque = ""
	req.URL.Path = sonarClient.BaseURL().Path + "issues/search"
	// use custom HTTP client to send request
	response, err = s.Client.Send(req)
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

func (s *Requester) decode(resp *http.Response, v interface{}) (err error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
