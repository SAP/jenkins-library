package trustengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
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
	defer response.Body.Close()
	if err != nil {
		// the body contains full error message which we want to log
		bodyBytes, err := io.ReadAll(response.Body)
		if err == nil {
			log.Entry().Info(string(bodyBytes))
		}
		return trust, err
	}

	err = json.NewDecoder(response.Body).Decode(&trust)
	if err != nil {
		return trust, err
	}
	return trust, nil
}
