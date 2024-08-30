package trustengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

type Secret struct {
	Token  string `json:"sonar,omitempty"`
	System string `json:"system,omitempty"`
}

type Response struct {
	Secrets []Secret
}

type Configuration struct {
	ServerURL string
	Token     string
}

func GetToken(refName string, client *piperhttp.Client, trustEngineConfiguration Configuration) (string, error) {
	secrets, err := GetSecrets([]string{refName}, client, trustEngineConfiguration)
	if err != nil {
		return "", err
	}
	for _, s := range secrets {
		if s.System == refName {
			return s.Token, nil
		}
	}
	return "", errors.New("could not find token in trust engine response")
}

func GetSecrets(refNames []string, client *piperhttp.Client, trustEngineConfiguration Configuration) ([]Secret, error) {
	var secrets []Secret
	response, err := GetResponse(refNames, client, trustEngineConfiguration)
	if err != nil {
		return secrets, err
	}
	for k, v := range response {
		secrets = append(secrets, Secret{
			System: k,
			Token:  v})
	}

	return secrets, nil
}

func GetResponse(refNames []string, client *piperhttp.Client, trustEngineConfiguration Configuration) (map[string]string, error) {
	var secrets map[string]string
	query := fmt.Sprintf("?systems=%s", strings.Join(refNames, ","))
	fullURL := trustEngineConfiguration.ServerURL + query

	log.Entry().Debugf("getting token from %s", fullURL)
	var header http.Header = map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", trustEngineConfiguration.Token)}}
	response, err := client.SendRequest(http.MethodGet, fullURL, nil, header, nil)
	if err != nil && response != nil {
		// the body contains full error message which we want to log
		bodyBytes, bodyErr := io.ReadAll(response.Body)
		if bodyErr == nil {
			log.Entry().Info(string(bodyBytes))
		}
	}
	if err != nil {
		return secrets, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&secrets)
	if err != nil {
		return secrets, err
	}

	return secrets, nil
}
