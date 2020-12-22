package cpi

import (
	"fmt"
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
	TokenURL, User, Pwd string
}

// GetBearerToken -Provides the bearer token for making CPI OData calls
func (tokenParameters TokenParameters) GetBearerToken() (string, error) {

	var testURL = "https://demo/oauth/token"
	// for supporting tests
	// with httpMockGcts we want to try only on actual odata API calls but not for OAuth token fetch calls
	// so we skip OAuth call for mock tests
	if tokenParameters.TokenURL == testURL {
		result := "demotoken"
		return result, nil
	}

	httpClient := &piperhttp.Client{}
	clientOptions := piperhttp.ClientOptions{
		Username: tokenParameters.User,
		Password: tokenParameters.Pwd,
	}
	httpClient.SetOptions(clientOptions)

	header := make(http.Header)
	header.Add("Accept", "application/json")
	tokenFinalURL := fmt.Sprintf("%s?grant_type=client_credentials", tokenParameters.TokenURL)
	method := "POST"
	resp, httpErr := httpClient.SendRequest(method, tokenFinalURL, nil, header, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	// for supporting tests
	// with httpMockGcts we want to try only on actual odata API calls but not for OAuth token fetch calls
	// so we pass Oauth token in advance and skip OAuth call for mock tests
	if resp.Header.Get("Authorization") != "" {
		result := resp.Header.Get("Authorization")
		return result, nil
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
	finalResult := jsonResponse.Path("access_token").Data().(string)
	return finalResult, nil
}
