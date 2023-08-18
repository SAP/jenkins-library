package xsuaa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const authHeaderKey = "Authorization"
const oneHourInSeconds = 3600.0

// XSUAA contains the fields to authenticate to a xsuaa service instance on BTP to retrieve a access token
// It also caches the latest retrieved access token
type XSUAA struct {
	OAuthURL        string
	ClientID        string
	ClientSecret    string
	CachedAuthToken AuthToken
}

// AuthToken provides a structure for the XSUAA auth token to be marshalled into
type AuthToken struct {
	TokenType   string        `json:"token_type"`
	AccessToken string        `json:"access_token"`
	ExpiresIn   time.Duration `json:"expires_in"`
	ExpiresAt   time.Time
}

// SetAuthHeaderIfNotPresent retrieves a XSUAA bearer token and sets the 'Authorization' header on a given http.Header.
// If another 'Authorization' header is already present, no change is done to the given header.
func (x *XSUAA) SetAuthHeaderIfNotPresent(header *http.Header) error {
	if len(header.Get(authHeaderKey)) > 0 {
		return nil
	}
	if len(x.OAuthURL) == 0 ||
		len(x.ClientID) == 0 ||
		len(x.ClientSecret) == 0 {
		return errors.Errorf("OAuthURL, ClientID and ClientSecret have to be set on the xsuaa instance")
	}

	secondsOfValidityLeft := x.CachedAuthToken.ExpiresAt.Sub(time.Now()).Seconds()
	if len(x.CachedAuthToken.AccessToken) == 0 ||
		(secondsOfValidityLeft > 0 && secondsOfValidityLeft < oneHourInSeconds) {
		token, err := x.GetBearerToken()
		if err != nil {
			return err
		}
		x.CachedAuthToken = token
	}
	header.Add(authHeaderKey, fmt.Sprintf("%s %s", x.CachedAuthToken.TokenType, x.CachedAuthToken.AccessToken))
	return nil
}

// GetBearerToken authenticates to and retrieves the auth information from the provided XSUAA oAuth base url. The following path
// and query is always used: /oauth/token?grant_type=client_credentials&response_type=token. The gotten JSON string is marshalled
// into an AuthToken struct and returned. If no 'access_token' field was present in the JSON response, an error is returned.
func (x *XSUAA) GetBearerToken() (authToken AuthToken, err error) {
	const method = http.MethodGet
	const urlPathAndQuery = "oauth/token?grant_type=client_credentials&response_type=token"

	oauthBaseURL, err := url.Parse(x.OAuthURL)
	if err != nil {
		return
	}
	entireURL := fmt.Sprintf("%s://%s/%s", oauthBaseURL.Scheme, oauthBaseURL.Host, urlPathAndQuery)

	httpClient := http.Client{}

	request, err := http.NewRequest(method, entireURL, nil)
	if err != nil {
		return
	}
	request.Header.Add("Accept", "application/json")
	request.SetBasicAuth(x.ClientID, x.ClientSecret)

	response, httpErr := httpClient.Do(request)
	if httpErr != nil {
		err = errors.Wrapf(httpErr, "fetching an access token failed: HTTP %s request to %s failed",
			method, entireURL)
		return
	}

	bodyText, err := readResponseBody(response)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		err = errors.Errorf("fetching an access token failed: HTTP %s request to %s failed: "+
			"expected response code 200, got '%d', response body: '%s'",
			method, entireURL, response.StatusCode, bodyText)
		return
	}

	parsingErr := json.Unmarshal(bodyText, &authToken)
	if err != nil {
		err = errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %s", bodyText)
		return
	}

	if authToken.AccessToken == "" {
		err = errors.Errorf("expected authToken field 'access_token' in json response: got response body: '%s'",
			bodyText)
		return
	}
	if authToken.TokenType == "" {
		authToken.TokenType = "bearer"
	}
	if authToken.ExpiresIn > 0 {
		authToken.ExpiresAt = setExpireTime(time.Now(), authToken.ExpiresIn)
	}

	return
}

func setExpireTime(now time.Time, secondsValid time.Duration) time.Time {
	return now.Add(time.Second * secondsValid)
}

func readResponseBody(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, errors.Errorf("did not retrieve an HTTP response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	bodyText, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, errors.Wrap(readErr, "HTTP response body could not be read")
	}
	return bodyText, nil
}
