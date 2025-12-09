package vault

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/vault/mocks"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestOIDC(t *testing.T) {
	oidcPath := "identity/oidc/token/testRoleID"
	mockPayload := base64.RawStdEncoding.EncodeToString([]byte("testOIDCtoken123"))
	mockToken := fmt.Sprintf("hvs.%s", mockPayload)

	mockJwt := &api.Secret{
		Data: map[string]interface{}{
			"path":  oidcPath,
			"token": mockToken,
		},
	}

	t.Run("Test getting OIDC token - token non-existent in env yet", func(t *testing.T) {
		t.Parallel()

		// init
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", oidcPath).Return(mockJwt, nil)

		// run
		token, err := client.GetOIDCTokenByValidation("testRoleID")

		// assert
		assert.NoError(t, err)
		assert.Equal(t, token, mockToken)
	})

	t.Run("Test getting OIDC token - token exists in env and is valid", func(t *testing.T) {
		// init
		// still valid for 10 minutes
		expiryTime := time.Now().Local().Add(time.Minute * time.Duration(10))
		payload := fmt.Sprintf("{\"exp\": %d}", expiryTime.Unix())
		payloadB64 := base64.RawStdEncoding.EncodeToString([]byte(payload))
		token := fmt.Sprintf("hvs.%s", payloadB64)

		t.Setenv("PIPER_OIDCIdentityToken", token)

		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", oidcPath).Return(mockJwt, nil)

		// run
		tokenResult, err := client.GetOIDCTokenByValidation("testRoleID")

		// assert
		assert.Equal(t, token, tokenResult)
		assert.NoError(t, err)
	})

	t.Run("Test getting OIDC token - token exists in env and is invalid", func(t *testing.T) {
		//init
		// expired 10 minutes ago (time is subtracted!)
		expiryTime := time.Now().Add(-time.Minute * time.Duration(10))
		payload := fmt.Sprintf("{\"exp\": %d}", expiryTime.Unix())
		payloadB64 := base64.RawStdEncoding.EncodeToString([]byte(payload))
		token := fmt.Sprintf("hvs.%s", payloadB64)

		t.Setenv("PIPER_OIDCIdentityToken", token)

		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", oidcPath).Return(mockJwt, nil)

		// run
		tokenResult, err := client.GetOIDCTokenByValidation("testRoleID")

		// assert
		client.GetOIDCTokenByValidation("testRoleID")
		assert.Equal(t, mockToken, tokenResult)
		assert.NoError(t, err)
	})

}
