package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/api/sts/v1"
)

// https://cloud.google.com/iam/docs/reference/sts/rest
const exchangeTokenAPIURL = "https://sts.googleapis.com/v1/token"

func GetFederatedToken(projectNumber, pool, provider, token string) (string, error) {
	ctx := context.Background()
	requestData := getExchangeTokenRequestData(projectNumber, pool, provider, token)

	// data to byte
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal the request data")
	}

	// build request
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, exchangeTokenAPIURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", errors.Wrap(err, "failed to build request")
	}

	// send request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request")
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code: %v", response.StatusCode)
	}

	// response to data
	defer response.Body.Close()
	responseData := sts.GoogleIdentityStsV1ExchangeTokenResponse{}
	err = json.NewDecoder(response.Body).Decode(&responseData)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode response")
	}

	return responseData.AccessToken, nil
}

func getExchangeTokenRequestData(projectNumber string, pool string, provider string, token string) sts.GoogleIdentityStsV1ExchangeTokenRequest {
	return sts.GoogleIdentityStsV1ExchangeTokenRequest{
		Audience: fmt.Sprintf(
			"//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s",
			projectNumber, pool, provider),
		Scope:              "https://www.googleapis.com/auth/cloud-platform",
		SubjectToken:       token,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:jwt",
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
	}
}
