package vault

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/vault/api"
)

type SecretData = map[string]interface{}

func TestGetKVSecret(t *testing.T) {
	t.Run("Test missing secret", func(t *testing.T) {
		client := Client{&mock.VaultClientMock{}}
		_, err := client.GetKVSecret("secret/notexist")
		assert.Error(t, err, "Expected to fail since the secret does not exist")
	})

	t.Run("Test parsing KV secrets", func(t *testing.T) {
		const secretName = "secret/supersecret"
		vaultMock := &mock.VaultClientMock{}
		client := Client{vaultMock}

		vaultMock.AddSecret(secretName, kvSecret(SecretData{"key1": "value1"}))
		secret, err := client.GetKVSecret(secretName)
		assert.NoError(t, err, "Excpect GetKVSecret to succeed")
		assert.Equal(t, "value1", secret["key1"])

		vaultMock.AddSecret(secretName, kvSecret(SecretData{"key1": "value1", "key2": 5}))
		secret, err = client.GetKVSecret(secretName)
		assert.Error(t, err, "Excpected to fail since value is wrong data type")

		vaultMock.AddSecret(secretName, &api.Secret{Data: SecretData{"key1": "value1"}})
		secret, err = client.GetKVSecret(secretName)
		assert.Error(t, err, "Expected to fail since 'data' field is missing")

		vaultMock.AddSecret(secretName, &api.Secret{Data: SecretData{"data": 5}})
		secret, err = client.GetKVSecret(secretName)
		assert.Error(t, err, "Expected to fail since 'data' has wrong type")

	})
}

func kvSecret(data SecretData) *api.Secret {
	return &api.Secret{
		Data: SecretData{"data": data},
	}
}
