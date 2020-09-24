package vault

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	mocks "github.com/SAP/jenkins-library/pkg/vault/mocks"

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
		client := Client{vaultMock}
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
			client := Client{vaultMock}
			vaultMock.On("Read", secretAPIPath).Return(kv2Secret(SecretData{"key1": "value1"}), nil)
			secret, err := client.GetKvSecret(secretName)
			assert.NoError(t, err, "Expect GetKvSecret to succeed")
			assert.Equal(t, "value1", secret["key1"])

		})

		t.Run("error is thrown when 'data' field can't be parsed", func(t *testing.T) {
			vaultMock := &mocks.VaultMock{}
			setupMockKvV2(vaultMock)
			client := Client{vaultMock}
			vaultMock.On("Read", secretAPIPath).Return(kv2Secret(SecretData{"key1": "value1", "key2": 5}), nil)
			secret, err := client.GetKvSecret(secretName)
			assert.Error(t, err, "Excpected to fail since value is wrong data type")
			assert.Nil(t, secret)

		})

		t.Run("error is thrown when data field is missing", func(t *testing.T) {
			vaultMock := &mocks.VaultMock{}
			setupMockKvV2(vaultMock)
			client := Client{vaultMock}
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
		client := Client{vaultMock}

		vaultMock.On("Read", mock.AnythingOfType("string")).Return(nil, nil)
		secret, err := client.GetKvSecret("secret/notexist")
		assert.NoError(t, err, "Missing secret should not an error")
		assert.Nil(t, secret)
	})

	t.Run("Test parsing KV1 secrets", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		setupMockKvV1(vaultMock)
		client := Client{vaultMock}

		vaultMock.On("Read", secretName).Return(kv1Secret(SecretData{"key1": "value1"}), nil)
		secret, err := client.GetKvSecret(secretName)
		assert.NoError(t, err)
		assert.Equal(t, "value1", secret["key1"])
	})

	t.Run("Test parsing KV1 secrets", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		setupMockKvV1(vaultMock)
		vaultMock.On("Read", secretName).Return(kv1Secret(SecretData{"key1": 5}), nil)
		client := Client{vaultMock}

		secret, err := client.GetKvSecret(secretName)
		assert.Error(t, err)
		assert.Nil(t, secret)

	})
}

func TestUnknownKvVersion(t *testing.T) {
	vaultMock := &mocks.VaultMock{}
	client := Client{vaultMock}

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
