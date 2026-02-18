package sonar

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	sonargo "github.com/magicsong/sonargo/sonar"
)

// Requester ...
type Requester struct {
	Client   Sender
	Host     string
	Username string
	Password string
	// TODO: implement certificate handling
	// Certificates [][]byte
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	Send(*http.Request) (*http.Response, error)
}

func (requester *Requester) create(method, path string, options any) (request *http.Request, err error) {
	sonarGoClient, err := sonargo.NewClient(requester.Host, requester.Username, requester.Password)
	if err != nil {
		return
	}
	// reuse request creation from sonargo
	request, err = sonarGoClient.NewRequest(method, path, options)
	if err != nil {
		return
	}
	// request created by sonarGO uses .Opaque without the host parameter leading to a request against https://api/issues/search
	// https://github.com/magicsong/sonargo/blob/103eda7abc20bd192a064b6eb94ba26329e339f1/sonar/sonarqube.go#L55
	request.URL.Opaque = ""
	request.URL.Path = sonarGoClient.BaseURL().Path + path
	return
}

func (requester *Requester) send(request *http.Request) (*http.Response, error) {
	return requester.Client.Send(request)
}

func (requester *Requester) decode(response *http.Response, result any) error {
	decoder := json.NewDecoder(response.Body)
	defer response.Body.Close()
	// sonargo.IssuesSearchObject does not imlement "internal" field organization and thus decoding fails
	// anyway the field is currently not needed so we simply allow (and drop) unknown fields to avoid extending the type
	// decoder.DisallowUnknownFields()
	return decoder.Decode(result)
}

// NewAPIClient ...
func NewAPIClient(host, token string, client Sender) *Requester {
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	// Make sure the given URL end with a api/
	if !strings.HasSuffix(host, "api/") {
		host += "api/"
	}
	log.Entry().Debugf("using api client for '%s'", host)
	return &Requester{
		Client:   client,
		Host:     host,
		Username: token,
	}
}
