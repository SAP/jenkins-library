package sonar

import (
	"encoding/json"
	"net/http"

	sonargo "github.com/magicsong/sonargo/sonar"
)

// Requester ...
type Requester struct {
	Client   Sender
	Host     string
	Username string
	Password string
	// Certificates [][]byte
	// CACert    []byte
	// SslVerify bool
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	Send(*http.Request) (*http.Response, error)
}

func (requester *Requester) create(method, path string, options interface{}) (request *http.Request, err error) {
	sonarGoClient, err := sonargo.NewClient(requester.Host, requester.Username, requester.Password)
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

func (requester *Requester) decode(response *http.Response, result interface{}) error {
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(result)
}
