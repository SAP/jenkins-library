package trustengine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

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
	ServerURL           string
	TokenEndPoint       string
	TokenQueryParamName string
	Token               string
}

// GetToken requests a single token
func GetToken(refName string, client *piperhttp.Client, trustEngineConfiguration Configuration) (string, error) {
	secrets, err := GetSecrets([]string{refName}, client, trustEngineConfiguration)
	if err != nil {
		return "", errors.Wrap(err, "couldn't get token from System Trust")
	}
	for _, s := range secrets {
		if s.System == refName {
			return s.Token, nil
		}
	}
	return "", errors.New("could not find token in System Trust response")
}

// GetSecrets transforms the System Trust JSON response into System Trust secrets, and can be used to request multiple tokens
func GetSecrets(refNames []string, client *piperhttp.Client, trustEngineConfiguration Configuration) ([]Secret, error) {
	var secrets []Secret
	query := url.Values{
		trustEngineConfiguration.TokenQueryParamName: {
			strings.Join(refNames, ","),
		},
	}
	response, err := getResponse(trustEngineConfiguration.ServerURL, trustEngineConfiguration.TokenEndPoint, query, client)
	if err != nil {
		return secrets, errors.Wrap(err, "getting secrets from System Trust failed")
	}
	for k, v := range response {
		secrets = append(secrets, Secret{
			System: k,
			Token:  v})
	}

	return secrets, nil
}

// getResponse returns a map of the JSON response that the System Trust puts out
func getResponse(serverURL, endpoint string, query url.Values, client *piperhttp.Client) (map[string]string, error) {
	var secrets map[string]string

	rawURL, err := parseURL(serverURL, endpoint, query)
	if err != nil {
		return secrets, errors.Wrap(err, "parsing System Trust url failed")
	}
	header := make(http.Header)
	header.Add("Accept", "application/json")

	log.Entry().Debugf("  with URL %s", rawURL)
	response, err := client.SendRequest(http.MethodGet, rawURL, nil, header, nil)
	if err != nil {
		if response != nil {
			// the body contains full error message which we want to log
			defer response.Body.Close()
			bodyBytes, bodyErr := io.ReadAll(response.Body)
			if bodyErr == nil {
				err = errors.Wrap(err, string(bodyBytes))
			}
		}
		return secrets, errors.Wrap(err, "getting response from System Trust failed")
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&secrets)
	if err != nil {
		return secrets, errors.Wrap(err, "getting response from System Trust failed")
	}

	return secrets, nil
}

// parseURL creates the full URL for a System Trust GET request
func parseURL(serverURL, endpoint string, query url.Values) (string, error) {
	rawFullEndpoint, err := url.JoinPath(serverURL, endpoint)
	if err != nil {
		return "", errors.New("error parsing System Trust URL")
	}
	fullURL, err := url.Parse(rawFullEndpoint)
	if err != nil {
		return "", errors.New("error parsing System Trust URL")
	}
	// commas and spaces shouldn't be escaped since the System Trust won't accept it
	unescapedRawQuery, err := url.QueryUnescape(query.Encode())
	if err != nil {
		return "", errors.New("error parsing System Trust URL")
	}
	fullURL.RawQuery = unescapedRawQuery
	return fullURL.String(), nil
}

// PrepareClient adds the System Trust authentication token to the client
func PrepareClient(client *piperhttp.Client, trustEngineConfiguration Configuration) *piperhttp.Client {
	var logEntry *logrus.Entry
	if logrus.GetLevel() < logrus.DebugLevel {
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		logEntry = logrus.NewEntry(logger)
	}
	client.SetOptions(piperhttp.ClientOptions{
		Token:  fmt.Sprintf("Bearer %s", trustEngineConfiguration.Token),
		Logger: logEntry,
	})
	return client
}
