package vault

import (
	"testing"

	"github.com/hashicorp/vault/api"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

type SecretData = map[string]interface{}

func TestGetKV2Secret(t *testing.T) {
	vaultMock := &mock.VaultClientMock{}

	t.Run("Test missing secret", func(t *testing.T) {
		client := Client{vaultMock}
		secret, err := client.GetKvSecret("secret/notexist")
		assert.NoError(t, err, "Missing secret should not an error")
		assert.Nil(t, secret)
	})

	t.Run("Test parsing KV2 secrets", func(t *testing.T) {
		const secretAPIPath = "secret/data/test"
		const secretName = "secret/test"
		client := Client{vaultMock}

		vaultMock.AddSecret(secretAPIPath, kv2Secret(SecretData{"key1": "value1"}))
		secret, err := client.GetKvSecret(secretName)
		assert.NoError(t, err, "Expect GetKvSecret to succeed")
		assert.Equal(t, "value1", secret["key1"])

		vaultMock.AddSecret(secretAPIPath, kv2Secret(SecretData{"key1": "value1", "key2": 5}))
		secret, err = client.GetKvSecret(secretName)
		assert.Error(t, err, "Excpected to fail since value is wrong data type")

		vaultMock.AddSecret(secretAPIPath, kv1Secret(SecretData{"key1": "value1"}))
		secret, err = client.GetKvSecret(secretName)
		assert.Error(t, err, "Expected to fail since 'data' field is missing")

		vaultMock.AddSecret(secretAPIPath, &api.Secret{Data: SecretData{"data": 5}})
		secret, err = client.GetKvSecret(secretName)
		assert.Error(t, err, "Expected to fail since 'data' has a wrong type")
	})
}

func TestGetKV1Secret(t *testing.T) {
	vaultMock := &mock.VaultClientMock{}
	vaultMock.SetKvEngineVersion(1)

	t.Run("Test missing secret", func(t *testing.T) {
		client := Client{vaultMock}
		secret, err := client.GetKvSecret("secret/notexist")
		assert.NoError(t, err, "Missing secret should not an error")
		assert.Nil(t, secret)
	})

	t.Run("Test parsing KV1 secrets", func(t *testing.T) {
		const secretAPIPath = "secret/data/test"
		const secretName = "secret/test"
		client := Client{vaultMock}

		vaultMock.AddSecret(secretName, kv1Secret(SecretData{"key1": "value1"}))
		secret, err := client.GetKvSecret(secretName)
		assert.NoError(t, err, "Expect GetKvSecret to succeed")
		assert.Equal(t, "value1", secret["key1"])

		vaultMock.AddSecret(secretName, kv1Secret(SecretData{"key1": "value1", "key2": 5}))
		secret, err = client.GetKvSecret(secretName)
		assert.Error(t, err, "Excpected to fail since value is wrong data type")

		vaultMock.AddSecret(secretName, &api.Secret{Data: SecretData{"data": 5}})
		secret, err = client.GetKvSecret(secretName)
		assert.Error(t, err, "Expected to fail since 'data' has wrong type")
	})
}

func TestUnknownKvVersion(t *testing.T) {
	vaultMock := &mock.VaultClientMock{}
	vaultMock.SetKvEngineVersion(3)
	client := Client{vaultMock}
	secret, err := client.GetKvSecret("/secret/secret")
	assert.EqualError(t, err, "KV Engine in version 3 is currently not supported")
	assert.Nil(t, secret)

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
