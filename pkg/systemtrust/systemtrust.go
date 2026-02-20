package systemtrust

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"errors"

	"github.com/sirupsen/logrus"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

type Secret struct {
	Token  string
	System string
}

type Configuration struct {
	ServerURL           string
	TokenEndPoint       string
	TokenQueryParamName string
	Token               string
}

type tokenRequestArray = []tokenRequest

type tokenRequest struct {
	System string `json:"system"`
	Scope  string `json:"scope"`
}

const defaultScope = "pipeline"

// GetToken requests a single token.
// By default, refName is used as the system and the default scope is applied.
// If refName contains "<scope>", the value before "<scope>" is used as the system
// and the value after "<scope>" is propagated as the scope.
func GetToken(refName string, client *piperhttp.Client, systemTrustConfiguration Configuration) (string, error) {
	body := refNameToTokenBody(refName)
	secrets, err := getSecrets(client, systemTrustConfiguration, body)
	if err != nil {
		return "", fmt.Errorf("couldn't get token from System Trust: %w", err)
	}
	for _, s := range secrets {
		if s.System == body.System {
			return s.Token, nil
		}
	}
	return "", errors.New("could not find token in System Trust response")
}

func refNameToTokenBody(refName string) tokenRequest {
	const marker = "<scope>"

	system := refName
	scope := defaultScope

	if strings.Contains(refName, marker) {
		parts := strings.SplitN(refName, marker, 2)

		if parts[0] != "" && parts[1] != "" {
			system = parts[0]
			scope = parts[1]
		}
	}

	return tokenRequest{
		System: system,
		Scope:  scope,
	}
}

// getSecrets using the system trust session token and convert to respectful system token based on request body
func getSecrets(client *piperhttp.Client, systemTrustConfiguration Configuration, requests ...tokenRequest) ([]Secret, error) {
	var secrets []Secret

	response, err := getResponse(systemTrustConfiguration.ServerURL, systemTrustConfiguration.TokenEndPoint, client, requests)
	if err != nil {
		return secrets, fmt.Errorf("getting secrets from System Trust failed: %w", err)
	}
	for k, v := range response {
		secrets = append(secrets, Secret{
			System: k,
			Token:  v})
	}

	return secrets, nil
}

// getResponse returns a map of the JSON response that the System Trust puts out
func getResponse(serverURL, endpoint string, client *piperhttp.Client, body tokenRequestArray) (map[string]string, error) {
	var secrets map[string]string

	rawURL, err := parseURL(serverURL, endpoint)
	if err != nil {
		return secrets, fmt.Errorf("parsing System Trust url failed: %w", err)
	}

	header := make(http.Header)
	header.Add("Accept", "application/json")

	bodyReader, err := trustTokenRequestToReader(body)
	if err != nil {
		return secrets, fmt.Errorf("failed to marshal token request body: %w", err)
	}

	log.Entry().Debugf("  with body %s", body)
	response, err := client.SendRequest(http.MethodPost, rawURL, bodyReader, header, nil)
	if err != nil {
		if response != nil {
			// the body contains full error message which we want to log
			defer response.Body.Close()
			bodyBytes, bodyErr := io.ReadAll(response.Body)
			if bodyErr == nil {
				err = fmt.Errorf(string(bodyBytes), err)
			}
		}
		return secrets, fmt.Errorf("getting response from System Trust failed: %w", err)
	}
	defer response.Body.Close()

	log.Entry().Debugf("  with response code %d", response.StatusCode)

	err = json.NewDecoder(response.Body).Decode(&secrets)
	if err != nil {
		return secrets, fmt.Errorf("getting response from System Trust failed: %w", err)
	}

	return secrets, nil
}

func trustTokenRequestToReader(body tokenRequestArray) (io.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request body: %w", err)
	}
	return bytes.NewReader(b), nil
}

// parseURL creates the full URL for a System Trust POST request
func parseURL(serverURL, endpoint string) (string, error) {
	rawFullEndpoint, err := url.JoinPath(serverURL, endpoint)
	if err != nil {
		return "", errors.New("error parsing System Trust URL")
	}
	fullURL, err := url.Parse(rawFullEndpoint)
	if err != nil {
		return "", errors.New("error parsing System Trust URL")
	}
	return fullURL.String(), nil
}

// PrepareClient adds the System Trust authentication token to the client
func PrepareClient(client *piperhttp.Client, systemTrustConfiguration Configuration) *piperhttp.Client {
	var logEntry *logrus.Entry
	if logrus.GetLevel() < logrus.DebugLevel {
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		logEntry = logrus.NewEntry(logger)
	}
	client.SetOptions(piperhttp.ClientOptions{
		Token:  fmt.Sprintf("Bearer %s", systemTrustConfiguration.Token),
		Logger: logEntry,
	})
	return client
}
