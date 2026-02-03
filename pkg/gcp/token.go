package gcp

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"google.golang.org/api/option"

	"github.com/pkg/errors"
	"google.golang.org/api/sts/v1"
)

const (
	gcpPubsubTokenKey       = "PIPER_gcpPubsubToken"
	gcpPubsubTokenExpiryKey = "PIPER_gcpPubsubTokenExpiresAt"
)

// carve out this part
// getFederatedToken tries to retrieve cached token from env variables, otherwise it will exchange
// OIDC identity token to access token and cache them in env variables
func getFederatedToken(projectNumber, pool, provider, oidcToken string) (string, error) {
	cachedToken := os.Getenv(gcpPubsubTokenKey)
	cachedExpiresAt := os.Getenv(gcpPubsubTokenExpiryKey)
	if tokenIsValid(cachedToken, cachedExpiresAt) {
		log.Entry().Debug("reusing GCP PubSub access token from cache")
		return cachedToken, nil
	}

	ctx := context.Background()
	token, expiresAt, err := exchangeOIDCToken(ctx, projectNumber, pool, provider, oidcToken)
	if err != nil {
		return "", errors.Wrap(err, "token exchange")
	}

	os.Setenv(gcpPubsubTokenKey, token)
	os.Setenv(gcpPubsubTokenExpiryKey, strconv.FormatInt(expiresAt, 10))
	return token, nil
}

// exchangeOIDCToken exchanges OIDC identity token to access token and returns expiry time in Unix timestamp
func exchangeOIDCToken(ctx context.Context, projectNumber, pool, provider, oidcToken string) (string, int64, error) {
	if len(oidcToken) == 0 {
		return "", 0, errors.New("OIDC identity token is absent")
	}

	stsService, err := sts.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		return "", 0, errors.Wrap(err, "service not created")
	}

	request := getExchangeTokenRequestData(projectNumber, pool, provider, oidcToken)
	response, err := sts.NewV1Service(stsService).Token(request).Context(ctx).Do()
	if err != nil {
		return "", 0, errors.Wrap(err, "exchange failed")
	}

	expiresAt := time.Now().Unix() + response.ExpiresIn
	log.Entry().Debugf("token successfully exchanged and will expire at %s", time.Unix(expiresAt, 0))
	return response.AccessToken, expiresAt, nil
}

func tokenIsValid(token string, expiresAtStr string) bool {
	if token == "" {
		return false
	}

	expiresAt, _ := strconv.Atoi(expiresAtStr)
	buffer := 5 // 5 second buffer to prevent using token that potentially may expire during execution
	if int64(expiresAt-buffer) < time.Now().Unix() {
		return false
	}

	return true
}

func getExchangeTokenRequestData(projectNumber string, pool string, provider string, token string) *sts.GoogleIdentityStsV1ExchangeTokenRequest {
	return &sts.GoogleIdentityStsV1ExchangeTokenRequest{
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
