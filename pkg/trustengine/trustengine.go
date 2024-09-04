package trustengine

import (
	"encoding/json"
	"errors"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Secret struct {
	Token  string
	System string
}

type Response struct {
	Secrets []Secret
}

type Configuration struct {
	ServerURL     string
	TokenEndPoint string
	Token         string
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
	endpoint := trustEngineConfiguration.TokenEndPoint
	queryValues := strings.Join(refNames, ",")
	response, err := getResponse(endpoint, queryValues, trustEngineConfiguration, client)
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

// getResponse returns a map of the JSON response that the trust engine puts out
func getResponse(endpoint, queryValues string, trustEngineConfiguration Configuration, client *piperhttp.Client) (map[string]string, error) {
	var secrets map[string]string
	fullURL, err := url.JoinPath(trustEngineConfiguration.ServerURL, endpoint)
	if err != nil {
		return secrets, errors.Join(err, errors.New("could not parse URL"))
	}
	fullURL = fullURL + queryValues

	header := make(http.Header)
	header.Add("Accept", "application/json")

	log.Entry().Debugf("with URL %s", fullURL)
	response, err := client.SendRequest(http.MethodGet, fullURL, nil, header, nil)
	if err != nil {
		if response != nil {
			// the body contains full error message which we want to log
			defer response.Body.Close()
			bodyBytes, bodyErr := io.ReadAll(response.Body)
			if bodyErr == nil {
				err = errors.Join(err, errors.New(string(bodyBytes)))
			}
		}
		return secrets, errors.Join(err, errors.New("getting response from trust engine failed"))
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&secrets)
	if err != nil {
		return secrets, errors.Join(err, errors.New("getting response from trust engine failed"))
	}

	return secrets, nil
}

// PrepareClient adds the Trust Engine authentication token to the client
func PrepareClient(client *piperhttp.Client, trustEngineConfiguration Configuration) *piperhttp.Client {
	client.SetOptions(piperhttp.ClientOptions{
		Token: fmt.Sprintf("Bearer %s", trustEngineConfiguration.Token),
	})
	return client
}
