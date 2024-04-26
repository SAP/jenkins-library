package vault

import (
	"encoding/base64"
	"testing"

	"github.com/SAP/jenkins-library/pkg/vault/mocks"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestOIDC(t *testing.T) {
	oidcPath := "identity/oidc/token/testRoleID"
	mockToken := base64.StdEncoding.EncodeToString([]byte("testOIDCtoken123"))

	mockJwt := &api.Secret{
		Data: map[string]interface{}{
			"path":  oidcPath,
			"token": mockToken,
		},
	}

	t.Run("Test getting OIDC token - token non-existent in env yet", func(t *testing.T) {
		t.Parallel()

		vaultMock := &mocks.VaultMock{}
		client := Client{vaultMock, &Config{}}
		vaultMock.On("Read", oidcPath).Return(mockJwt, nil)

		token, err := client.GetOIDCTokenByValidation("testRoleID")

		assert.NoError(t, err)
		assert.Equal(t, token, mockToken)
	})

	// t.Run("Test getting OIDC token - token exists in env and is valid", func(t *testing.T) {
	// 	t.Parallel()

	// 	t.Setenv("PIPER_OIDCIdentityToken", "testOIDCtoken123")

	// 	vaultMock := &mocks.VaultMock{}
	// 	_ = Client{vaultMock, &Config{}}
	// 	vaultMock.On("Read", oidcPath).Return(mockJwt, nil)
	// })

	// t.Run("Test getting OIDC token - token exists in env and is invalid", func(t *testing.T) {
	// 	t.Parallel()

	// 	t.Setenv("PIPER_OIDCIdentityToken", "testOIDCtoken123")

	// 	vaultMock := &mocks.VaultMock{}
	// 	_ = Client{vaultMock, &Config{}}
	// 	vaultMock.On("Read", oidcPath).Return(mockJwt, nil)
	// })

}
