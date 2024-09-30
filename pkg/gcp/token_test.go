package gcp

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

//const exchangeTokenAPIURL = "https://sts.googleapis.com/v1/token"
//
//func TestGetFederatedToken(t *testing.T) {
//	t.Run("success", func(t *testing.T) {
//		// init
//		projectID := "PROJECT_NUMBER"
//		pool := "POOL"
//		provider := "PROVIDER"
//
//		// mock
//		httpmock.Activate()
//		defer httpmock.DeactivateAndReset()
//		httpmock.RegisterResponder(http.MethodPost, exchangeTokenAPIURL,
//			func(req *http.Request) (*http.Response, error) {
//				return httpmock.NewJsonResponse(http.StatusOK, sts.GoogleIdentityStsV1ExchangeTokenResponse{AccessToken: mock.Anything})
//			},
//		)
//
//		// test
//		federatedToken, err := GetFederatedToken(projectID, pool, provider, mock.Anything)
//		// asserts
//		assert.NoError(t, err)
//		assert.Equal(t, mock.Anything, federatedToken)
//	})
//}

func Test_tokenIsValid(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		expiresAtStr string
		want         bool
	}{
		{
			"token is empty",
			"",
			"",
			false,
		}, {
			"token expiredAt is empty",
			"someToken",
			"",
			false,
		}, {
			"token is expired",
			"someToken",
			fmt.Sprintf("%d", time.Now().Unix()-100),
			false,
		}, {
			"token is valid",
			"someToken",
			fmt.Sprintf("%d", time.Now().Unix()+100),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tokenIsValid(tt.token, tt.expiresAtStr), "tokenIsValid(%v, %v)", tt.token, tt.expiresAtStr)
		})
	}
}
