package ans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

// ANS holds the setup for the xsuaa service to retrieve a bearer token for authorization and
// the URL to the SAP Alert Notification Service backend
type ANS struct {
	XSUAA xsuaa.XSUAA
	URL   string
}

// Client to send the event to the SAP Alert Notification Service
type Client interface {
	Send(event Event) error
}

// ServiceKey holds the information about the SAP Alert Notification Service to send the events to
type ServiceKey struct {
	Url          string `json:"url"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	OauthUrl     string `json:"oauth_url"`
}

// UnmarshallServiceKeyJSON unmarshalls the given json service key string.
func UnmarshallServiceKeyJSON(serviceKeyJSON string) (ansServiceKey ServiceKey, err error) {
	err = json.Unmarshal([]byte(serviceKeyJSON), &ansServiceKey)
	if err != nil {
		err = errors.Wrap(err, "error unmarshalling ANS serviceKey")
		return
	}
	return
}

// Send an event to the SAP Alert Notification Service
func (ans ANS) Send(event Event) error {
	const eventPath = "/cf/producer/v1/resource-events"

	requestBody, err := json.Marshal(event)
	if err != nil {
		return err
	}

	header := make(http.Header)
	err = ans.XSUAA.SetAuthHeaderIfNotPresent(&header)
	if err != nil {
		return err
	}

	entireUrl := ans.URL + eventPath

	httpClient := http.Client{}
	request, err := http.NewRequest(http.MethodPost, entireUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	request.Header.Add(authHeaderKey, header.Get(authHeaderKey))
	request.Header.Add("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusAccepted {
		statusCodeError := fmt.Errorf("ANS http request to '%s' failed. Did not get expected status code %d; instead got %d",
			entireUrl, http.StatusAccepted, response.StatusCode)
		responseBody, err := readResponseBody(response)
		if err != nil {
			err = errors.Wrapf(err, "%s; reading response body failed", statusCodeError.Error())
		} else {
			err = fmt.Errorf("%s; response body: %s", statusCodeError.Error(), responseBody)
		}
		return err
	}

	return nil
}

func readResponseBody(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, errors.Errorf("did not retrieve an HTTP response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	bodyText, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, errors.Wrap(readErr, "HTTP response body could not be read")
	}
	return bodyText, nil
}
