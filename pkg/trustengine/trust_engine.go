package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"net/http"
)

type trustEngineUtils interface {
	*piperhttp.Client
}

type TrustEngineSecret struct {
	Token string `json:"sonar,omitempty"`
}

type TrustEngineConfiguration struct {
	ServerURL string
	Token     string
}

func GetTrustEngineSecret(refName string, client *piperhttp.Client, trustEngineConfiguration TrustEngineConfiguration) (string, error) {
	secret, err := GetTrustEngineResponse(refName, client, trustEngineConfiguration)
	if err != nil {
		return "", err
	}

	token := secret.Token
	if token == "" {
		return "", errors.New("no token found in trust engine response")
	}
	return token, nil
}

func GetTrustEngineResponse(refName string, client *piperhttp.Client, trustEngineConfiguration TrustEngineConfiguration) (TrustEngineSecret, error) {
	var trust TrustEngineSecret
	fullURL := trustEngineConfiguration.ServerURL + fmt.Sprintf("?systems=%s", refName)

	log.Entry().Debugf("getting token from %s", fullURL)
	var header http.Header = map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", trustEngineConfiguration.Token)}}
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
