//go:build unit
// +build unit

package vault

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/mock"

	"github.com/SAP/jenkins-library/pkg/vault/mocks"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

type SecretData = map[string]interface{}

const (
	sysLookupPath = "sys/internal/ui/mounts/"
)

func TestGetKV2Secret(t *testing.T) {
	t.Run("Test missing secret", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		setupMockKvV2(vaultMock)
		vaultMock.On("Read", "secret/data/notexist").Return(nil, nil)
		secret, err := client.GetKvSecret("secret/notexist")
		assert.NoError(t, err, "Missing secret should not an error")
		assert.Nil(t, secret)
	})

	t.Run("Test parsing KV2 secrets", func(t *testing.T) {
		t.Parallel()
		const secretAPIPath = "secret/data/test"
		const secretName = "secret/test"
		t.Run("Getting secret from KV engine (v2)", func(t *testing.T) {
			vaultMock := &mocks.VaultMock{}
			setupMockKvV2(vaultMock)
			client := Client{nil, vaultMock, &ClientConfig{}}
			vaultMock.On("Read", secretAPIPath).Return(kv2Secret(SecretData{"key1": "value1", "key2": map[string]any{"subkey1": "subvalue2"}, "key3": 3, "key4": []string{"foo", "bar"}}), nil)
			secret, err := client.GetKvSecret(secretName)
			assert.NoError(t, err, "Expect GetKvSecret to succeed")
			assert.Equal(t, "value1", secret["key1"])
			assert.Equal(t, `{"subkey1":"subvalue2"}`, secret["key2"])
			assert.Equal(t, "3", secret["key3"])
			assert.Equal(t, `["foo","bar"]`, secret["key4"])
		})

		t.Run("error is thrown when data field is missing", func(t *testing.T) {
			vaultMock := &mocks.VaultMock{}
			setupMockKvV2(vaultMock)
			client := Client{nil, vaultMock, &ClientConfig{}}
			vaultMock.On("Read", secretAPIPath).Return(kv1Secret(SecretData{"key1": "value1"}), nil)
			secret, err := client.GetKvSecret(secretName)
			assert.Error(t, err, "Expected to fail since 'data' field is missing")
			assert.Nil(t, secret)
		})
	})
}

func TestGetKV1Secret(t *testing.T) {
	t.Parallel()
	const secretName = "secret/test"

	t.Run("Test missing secret", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		setupMockKvV1(vaultMock)
		client := Client{nil, vaultMock, &ClientConfig{}}

		vaultMock.On("Read", mock.AnythingOfType("string")).Return(nil, nil)
		secret, err := client.GetKvSecret("secret/notexist")
		assert.NoError(t, err, "Missing secret should not an error")
		assert.Nil(t, secret)
	})

	t.Run("Test parsing KV1 secrets", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		setupMockKvV1(vaultMock)
		client := Client{nil, vaultMock, &ClientConfig{}}

		vaultMock.On("Read", secretName).Return(kv1Secret(SecretData{"key1": "value1", "key2": 5}), nil)
		secret, err := client.GetKvSecret(secretName)
		assert.NoError(t, err)
		assert.Equal(t, "value1", secret["key1"])
		assert.Equal(t, "5", secret["key2"])
	})
}

func TestSecretIDGeneration(t *testing.T) {
	t.Parallel()
	const secretID = "secret-id"
	const appRoleName = "test"
	const appRolePath = "auth/approle/role/test"

	t.Run("Test generating new secret-id", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		now := time.Now()
		expiry := now.Add(5 * time.Hour).Format(time.RFC3339)
		metadata := map[string]interface{}{
			"field1": "value1",
		}

		metadataJSON, err := json.Marshal(metadata)
		assert.NoError(t, err)
		vaultMock.On("Write", path.Join(appRolePath, "secret-id/lookup"), mapWith("secret_id", secretID)).Return(kv1Secret(SecretData{
			"expiration_time": expiry,
			"metadata":        metadata,
		}), nil)

		vaultMock.On("Write", path.Join(appRolePath, "/secret-id"), mapWith("metadata", string(metadataJSON))).Return(kv1Secret(SecretData{
			"secret_id": "newSecretId",
		}), nil)

		newSecretID, err := client.GenerateNewAppRoleSecret(secretID, appRoleName)
		assert.NoError(t, err)
		assert.Equal(t, "newSecretId", newSecretID)
	})

	t.Run("Test with no secret-id returned", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		now := time.Now()
		expiry := now.Add(5 * time.Hour).Format(time.RFC3339)
		metadata := map[string]interface{}{
			"field1": "value1",
		}

		metadataJSON, err := json.Marshal(metadata)
		assert.NoError(t, err)
		vaultMock.On("Write", path.Join(appRolePath, "secret-id/lookup"), mapWith("secret_id", secretID)).Return(kv1Secret(SecretData{
			"expiration_time": expiry,
			"metadata":        metadata,
		}), nil)

		vaultMock.On("Write", path.Join(appRolePath, "/secret-id"), mapWith("metadata", string(metadataJSON))).Return(kv1Secret(SecretData{}), nil)

		newSecretID, err := client.GenerateNewAppRoleSecret(secretID, appRoleName)
		assert.EqualError(t, err, fmt.Sprintf("Vault response for path %s did not contain a new secret-id", path.Join(appRolePath, "secret-id")))
		assert.Equal(t, newSecretID, "")
	})

	t.Run("Test with no new secret-id returned", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		now := time.Now()
		expiry := now.Add(5 * time.Hour).Format(time.RFC3339)
		metadata := map[string]interface{}{
			"field1": "value1",
		}

		metadataJSON, err := json.Marshal(metadata)
		assert.NoError(t, err)
		vaultMock.On("Write", path.Join(appRolePath, "secret-id/lookup"), mapWith("secret_id", secretID)).Return(kv1Secret(SecretData{
			"expiration_time": expiry,
			"metadata":        metadata,
		}), nil)

		vaultMock.On("Write", path.Join(appRolePath, "/secret-id"), mapWith("metadata", string(metadataJSON))).Return(kv1Secret(nil), nil)

		newSecretID, err := client.GenerateNewAppRoleSecret(secretID, appRoleName)
		assert.EqualError(t, err, fmt.Sprintf("could not generate new approle secret-id for approle path %s", path.Join(appRolePath, "secret-id")))
		assert.Equal(t, newSecretID, "")
	})
}

func TestSecretIDTtl(t *testing.T) {
	t.Parallel()
	const secretID = "secret-id"
	const appRolePath = "auth/approle/role/test"
	const appRoleName = "test"

	t.Run("Test fetching secreID TTL", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		now := time.Now()
		expiry := now.Add(5 * time.Hour).Format(time.RFC3339)
		vaultMock.On("Write", path.Join(appRolePath, "secret-id/lookup"), mapWith("secret_id", secretID)).Return(kv1Secret(SecretData{
			"expiration_time": expiry,
		}), nil)

		ttl, err := client.GetAppRoleSecretIDTtl(secretID, appRoleName)
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Hour, ttl.Round(time.Hour))
	})

	t.Run("Test with no expiration time", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Write", path.Join(appRolePath, "secret-id/lookup"), mapWith("secret_id", secretID)).Return(kv1Secret(SecretData{}), nil)
		ttl, err := client.GetAppRoleSecretIDTtl(secretID, appRoleName)
		assert.EqualError(t, err, fmt.Sprintf("could not load secret-id information from path %s", appRolePath))
		assert.Equal(t, time.Duration(0), ttl)
	})

	t.Run("Test with wrong date format", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Write", path.Join(appRolePath, "secret-id/lookup"), mapWith("secret_id", secretID)).Return(kv1Secret(SecretData{
			"expiration_time": time.Now().String(),
		}), nil)
		ttl, err := client.GetAppRoleSecretIDTtl(secretID, appRoleName)
		assert.True(t, strings.HasPrefix(err.Error(), "parsing time"))
		assert.Equal(t, time.Duration(0), ttl)
	})

	t.Run("Test with expired secret-id", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		now := time.Now()
		expiry := now.Add(-5 * time.Hour).Format(time.RFC3339)
		vaultMock.On("Write", path.Join(appRolePath, "secret-id/lookup"), mapWith("secret_id", secretID)).Return(kv1Secret(SecretData{
			"expiration_time": expiry,
		}), nil)

		ttl, err := client.GetAppRoleSecretIDTtl(secretID, appRoleName)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), ttl)
	})
}

func TestGetAppRoleName(t *testing.T) {
	t.Parallel()
	const secretID = "secret-id"

	t.Run("Test that correct role name is returned", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", "auth/token/lookup-self").Return(kv1Secret(SecretData{
			"meta": SecretData{
				"role_name": "test",
			},
		}), nil)

		appRoleName, err := client.GetAppRoleName()
		assert.NoError(t, err)
		assert.Equal(t, "test", appRoleName)
	})

	t.Run("Test without secret data", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", "auth/token/lookup-self").Return(kv1Secret(nil), nil)

		appRoleName, err := client.GetAppRoleName()
		assert.EqualError(t, err, "could not lookup token information: auth/token/lookup-self")
		assert.Empty(t, appRoleName)
	})

	t.Run("Test without metadata data", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", "auth/token/lookup-self").Return(kv1Secret(SecretData{}), nil)

		appRoleName, err := client.GetAppRoleName()
		assert.EqualError(t, err, "token info did not contain metadata auth/token/lookup-self")
		assert.Empty(t, appRoleName)
	})

	t.Run("Test without role name in metadata", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", "auth/token/lookup-self").Return(kv1Secret(SecretData{
			"meta": SecretData{},
		}), nil)

		appRoleName, err := client.GetAppRoleName()
		assert.Empty(t, appRoleName)
		assert.NoError(t, err)
	})

	t.Run("Test that different role_name types are ignored", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Read", "auth/token/lookup-self").Return(kv1Secret(SecretData{
			"meta": SecretData{
				"role_name": 5,
			},
		}), nil)

		appRoleName, err := client.GetAppRoleName()
		assert.Empty(t, appRoleName)
		assert.NoError(t, err)
	})
}

func TestTokenRevocation(t *testing.T) {
	t.Parallel()
	t.Run("Test that revocation error is returned", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Write",
			"auth/token/revoke-self",
			mock.IsType(map[string]interface{}{})).Return(nil, errors.New("Test"))

		err := client.RevokeToken()
		assert.Errorf(t, err, "Test")
	})

	t.Run("Test that revocation endpoint is called", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		client := Client{nil, vaultMock, &ClientConfig{}}
		vaultMock.On("Write",
			"auth/token/revoke-self",
			mock.IsType(map[string]interface{}{})).Return(nil, nil)
		err := client.RevokeToken()
		assert.NoError(t, err)
	})
}

func TestUnknownKvVersion(t *testing.T) {
	vaultMock := &mocks.VaultMock{}
	client := Client{nil, vaultMock, &ClientConfig{}}

	vaultMock.On("Read", "sys/internal/ui/mounts/secret/secret").Return(&api.Secret{
		Data: map[string]interface{}{
			"path": "secret",
			"options": map[string]interface{}{
				"version": "3",
			},
		}}, nil)

	secret, err := client.GetKvSecret("/secret/secret")
	assert.EqualError(t, err, "KV Engine in version 3 is currently not supported")
	assert.Nil(t, secret)

}

func setupMockKvV2(vaultMock *mocks.VaultMock) {
	vaultMock.On("Read", mock.MatchedBy(func(path string) bool {
		return strings.HasPrefix(path, sysLookupPath)
	})).Return(func(path string) *api.Secret {
		pathComponents := strings.Split(strings.TrimPrefix(path, "sys/internal/ui/mounts/"), "/")
		mountpath := "/"
		if len(pathComponents) > 1 {
			mountpath = pathComponents[0]
		}
		return &api.Secret{
			Data: map[string]interface{}{
				"path": mountpath,
				"options": map[string]interface{}{
					"version": "2",
				},
			},
		}
	}, nil)
}

func setupMockKvV1(vaultMock *mocks.VaultMock) {
	vaultMock.On("Read", mock.MatchedBy(func(path string) bool {
		return strings.HasPrefix(path, sysLookupPath)
	})).Return(func(path string) *api.Secret {
		pathComponents := strings.Split(strings.TrimPrefix(path, "sys/internal/ui/mounts/"), "/")
		mountpath := "/"
		if len(pathComponents) > 1 {
			mountpath = pathComponents[0]
		}
		return &api.Secret{
			Data: map[string]interface{}{
				"path": mountpath,
			},
		}
	}, nil)
}

func kv1Secret(data SecretData) *api.Secret {
	return &api.Secret{
		Data: data,
	}
}

func kv2Secret(data SecretData) *api.Secret {
	return &api.Secret{
		Data: SecretData{"data": data},
	}
}

func mapWith(key, expectedValue string) interface{} {
	return mock.MatchedBy(func(arg map[string]interface{}) bool {
		valRaw, ok := arg[key]
		if !ok {
			return false
		}

		val, ok := valRaw.(string)
		if !ok {
			return false
		}

		return val == expectedValue
	})
}
