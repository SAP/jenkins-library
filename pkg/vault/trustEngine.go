package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"net/http"
	"net/url"

	"github.com/SAP/jenkins-library/pkg/log"
)

type trustEngineUtils interface {
	*piperhttp.Client
}

type TrustEngineSecret struct {
	Token string `json:"sonar,omitempty"`
}

func GetTrustEngineSecret(baseURL *url.URL, refName, jwt string, client *piperhttp.Client) (string, error) {
	secret, err := GetTrustEngineResponse(baseURL, refName, jwt, client)
	if err != nil {
		return "", err
	}

	token := secret.Token
	if token == "" {
		return "", errors.New("no token found in trust engine response")
	}
	return token, nil
}

func GetTrustEngineResponse(baseURL *url.URL, refName, jwt string, client *piperhttp.Client) (TrustEngineSecret, error) {
	var trust TrustEngineSecret
	fullURL := baseURL.String() + fmt.Sprintf("?systems=%s", refName)

	log.Entry().Debugf("getting token from %s", fullURL)
	var header http.Header = map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", jwt)}}
	response, err := client.SendRequest("GET", fullURL, nil, header, nil)
	if err != nil {
		// is the full error message that the API returns being logged?
		return trust, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&trust)
	if err != nil {
		return trust, err
	}
	return trust, nil
}
