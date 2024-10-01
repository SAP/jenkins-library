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
			fmt.Sprintf("%d", time.Now().Unix()-100), // expiresAt is 100 seconds ahead
			false,
		}, {
			"token is valid",
			"someToken",
			fmt.Sprintf("%d", time.Now().Unix()+100), // expiresAt is 100 seconds before
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tokenIsValid(tt.token, tt.expiresAtStr), "tokenIsValid(%v, %v)", tt.token, tt.expiresAtStr)
		})
	}
}
