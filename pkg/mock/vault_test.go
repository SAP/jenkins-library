package mock

import (
	"testing"

	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestStoreSecret(t *testing.T) {
	t.Parallel()
	t.Run("no init", func(t *testing.T) {
		vault := VaultClientMock{}
		secret, err := vault.Read("/secret/test")
		assert.NoError(t, err)
		assert.Nil(t, secret)
	})

	t.Run("Secret exists after AddSecret", func(t *testing.T) {
		vault := VaultClientMock{}
		vault.AddSecret("/secret/test", &api.Secret{Data: map[string]interface{}{"key1": "value1"}})
		secret, err := vault.Read("/secret/test")
		assert.NoError(t, err)
		assert.Equal(t, "value1", secret.Data["key1"])
	})

	t.Run("Test KV Version response", func(t *testing.T) {
		vault := VaultClientMock{}

		secret, err := vault.Read("sys/internal/ui/mounts/secret/test")
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, "secret", secret.Data["path"])
		options := secret.Data["options"]
		assert.NotNil(t, options)
		optionsMap, ok := options.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "2", optionsMap["version"])

		vault.SetKvEngineVersion(1)
		secret, err = vault.Read("sys/internal/ui/mounts/secret/test")
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, "secret", secret.Data["path"])
		assert.Nil(t, secret.Data["options"])

	})
}
