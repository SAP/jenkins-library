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
	Token  string
	System string
}

type Response struct {
	Secrets []Secret
}

type Configuration struct {
	ServerURL string
	Token     string
}

// GetToken requests a single token
func GetToken(refName string, client *piperhttp.Client, trustEngineConfiguration Configuration) (string, error) {
	secrets, err := GetSecrets([]string{refName}, client, trustEngineConfiguration)
	if err != nil {
		return "", errors.Join(err, errors.New("couldn't get token from trust engine"))
	}
	for _, s := range secrets {
		if s.System == refName {
			return s.Token, nil
		}
	}
	return "", errors.New("could not find token in trust engine response")
}

// GetSecrets transforms the trust engine JSON response into trust engine secrets, and can be used to request multiple tokens
func GetSecrets(refNames []string, client *piperhttp.Client, trustEngineConfiguration Configuration) ([]Secret, error) {
	var secrets []Secret
	response, err := GetResponse(refNames, client, trustEngineConfiguration)
	if err != nil {
		return secrets, errors.Join(err, errors.New("getting secrets from trust engine failed"))
	}
	for k, v := range response {
		secrets = append(secrets, Secret{
			System: k,
			Token:  v})
	}

	return secrets, nil
}

// GetResponse returns a map of the JSON response that the trust engine puts out
func GetResponse(refNames []string, client *piperhttp.Client, trustEngineConfiguration Configuration) (map[string]string, error) {
	var secrets map[string]string
	query := fmt.Sprintf("?systems=%s", strings.Join(refNames, ","))
	fullURL := trustEngineConfiguration.ServerURL + query

	client.SetOptions(piperhttp.ClientOptions{
		Token: fmt.Sprintf("Bearer %s", trustEngineConfiguration.Token),
	})
	header := make(http.Header)
	header.Add("Accept", "application/json")

	log.Entry().Debugf("with URL %s", fullURL)
	response, err := client.SendRequest(http.MethodGet, fullURL, nil, header, nil)
	if err != nil && response != nil {
		// the body contains full error message which we want to log
		bodyBytes, bodyErr := io.ReadAll(response.Body)
		if bodyErr == nil {
			err = errors.Join(err, errors.New(string(bodyBytes)))
		}
	}
	if err != nil {
		return secrets, errors.Join(err, errors.New("getting response from trust engine failed"))
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&secrets)
	if err != nil {
		return secrets, errors.Join(err, errors.New("getting response from trust engine failed"))
	}

	return secrets, nil
}
