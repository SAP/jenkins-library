package gcp

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/sts/v1"
)

func TestGetExchangeTokenRequestData(t *testing.T) {
	// ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		// init
		projectNumber := "PROJECT_NUMBER"
		pool := "POOL"
		provider := "PROVIDER"
		// test
		data := getExchangeTokenRequestData(projectNumber, pool, provider, mock.Anything)
		// asserts
		assert.Equal(t, data.Audience, "//iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL/providers/PROVIDER")
		assert.Equal(t, data.SubjectToken, mock.Anything)
	})
}

func TestGetFederatedToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// init
		projectNumber := "PROJECT_NUMBER"
		pool := "POOL"
		provider := "PROVIDER"

		// mock
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodPost, exchangeTokenAPIURL,
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewJsonResponse(http.StatusOK, sts.GoogleIdentityStsV1ExchangeTokenResponse{AccessToken: mock.Anything})
			},
		)

		// test
		federatedToken, err := GetFederatedToken(projectNumber, pool, provider, mock.Anything)
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, mock.Anything, federatedToken)
	})
}
