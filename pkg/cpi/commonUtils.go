package cpi

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"io/ioutil"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
)

//CommonUtils for CPI
type CommonUtils interface {
	GetBearerToken() (string, error)
}

//TokenParameters struct
type TokenParameters struct {
	TokenURL, Username, Password string
	Client                       piperhttp.Sender
}

// CpiServiceKey contains information about a CPI service key
type CpiServiceKey struct {
	Host string `json:"url"`
	Uaa  OAuth  `json:"uaa"`
}

type OAuth struct {
	OAuthTokenProviderURL string `json:"url"`
	ClientId              string `json:"clientid"`
	ClientSecret          string `json:"clientsecret"`
}

// ReadCpiServiceKeyFile unmarshalls the give json service key string.
func ReadCpiServiceKeyFile(serviceKeyPath string, fileUtils piperutils.FileUtils) (cpiServiceKey CpiServiceKey, err error) {

	serviceKeyJSON, err := fileUtils.FileRead(serviceKeyPath)
	if err != nil {
		err = errors.Wrap(err, "error reading serviceKey file")
		return
	}

	// parse
	err = json.Unmarshal([]byte(serviceKeyJSON), &cpiServiceKey)
	if err != nil {
		err = errors.Wrap(err, "error unmarshalling serviceKey")
		return
	}

	log.Entry().Info("CPI serviceKey read successfully")
	return
}

// GetBearerToken -Provides the bearer token for making CPI OData calls
func (tokenParameters TokenParameters) GetBearerToken() (string, error) {

	httpClient := tokenParameters.Client

	clientOptions := piperhttp.ClientOptions{
		Username: tokenParameters.Username,
		Password: tokenParameters.Password,
	}
	httpClient.SetOptions(clientOptions)

	header := make(http.Header)
	header.Add("Accept", "application/json")
	tokenFinalURL := fmt.Sprintf("%s?grant_type=client_credentials", tokenParameters.TokenURL)
	method := "POST"
	resp, httpErr := httpClient.SendRequest(method, tokenFinalURL, nil, header, nil)
	if httpErr != nil {
		return "", errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", method, tokenFinalURL)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response")
	}

	if resp.StatusCode != 200 {
		return "", errors.Errorf("did not retrieve a valid HTTP response code: %v", httpErr)
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return "", errors.Wrap(readErr, "HTTP response body could not be read")
	}
	jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
	if parsingErr != nil {
		return "", errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
	}
	token := jsonResponse.Path("access_token").Data().(string)
	return token, nil
}
