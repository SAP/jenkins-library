package sonar

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/google/go-querystring/query"
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

func (requester *Requester) create(method, path string, options any) (*http.Request, error) {
	requestURL, err := url.Parse(requester.Host + path)
	if err != nil {
		return nil, err
	}
	if options != nil {
		values, err := query.Values(options)
		if err != nil {
			return nil, err
		}
		requestURL.RawQuery = values.Encode()
	}
	request, err := http.NewRequest(method, requestURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	if method == http.MethodPost || method == http.MethodPut {
		request.Header.Set("Content-Type", "application/json")
	}
	request.SetBasicAuth(requester.Username, requester.Password)
	return request, nil
}

func (requester *Requester) send(request *http.Request) (*http.Response, error) {
	return requester.Client.Send(request)
}

func (requester *Requester) decode(response *http.Response, result any) error {
	decoder := json.NewDecoder(response.Body)
	defer response.Body.Close()
	// IssuesSearchObject does not implement the "internal" field organization and thus decoding fails
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
