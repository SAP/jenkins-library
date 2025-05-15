package contrast

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

type ContrastHttpClient interface {
	ExecuteRequest(url string, params map[string]string, dest interface{}) error
}

type ContrastHttpClientInstance struct {
	apiKey string
	auth   string
}

func NewContrastHttpClient(apiKey, auth string) *ContrastHttpClientInstance {
	return &ContrastHttpClientInstance{
		apiKey: apiKey,
		auth:   auth,
	}
}

func (c *ContrastHttpClientInstance) ExecuteRequest(url string, params map[string]string, dest interface{}) error {
	req, err := newHttpRequest(url, c.apiKey, c.auth, params)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	log.Entry().Debugf("GET call request to: %s", url)
	response, err := performRequest(req)
	if response != nil && response.StatusCode != http.StatusOK {
		return errors.Errorf("failed to perform request, status code: %v and status %v", response.StatusCode, response.Status)
	}

	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	defer response.Body.Close()
	err = parseJsonResponse(response, dest)
	if err != nil {
		return errors.Wrap(err, "failed to parse JSON response")
	}
	return nil
}

func newHttpRequest(url, apiKey, auth string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("API-Key", apiKey)
	req.Header.Add("Authorization", auth)
	q := req.URL.Query()
	for param, value := range params {
		q.Add(param, value)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}
func performRequest(req *http.Request) (*http.Response, error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func parseJsonResponse(response *http.Response, jsonData interface{}) error {
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, jsonData)
	if err != nil {
		return err
	}
	return nil
}
