//go:build unit
// +build unit

package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockVaultRotateSecretIDUtilsBundle struct {
	t                *testing.T
	newSecret        string
	ttl              time.Duration
	config           *vaultRotateSecretIdOptions
	updateFuncCalled bool
}

func TestRunVaultRotateSecretId(t *testing.T) {
	t.Parallel()
	mock := &mockVaultRotateSecretIDUtilsBundle{t, "test-secret", time.Hour, getTestConfig(), false}
	runVaultRotateSecretID(mock)
	assert.True(t, mock.updateFuncCalled)

}

func TestRunVaultRotateSecretID(t *testing.T) {
	t.Parallel()

	t.Run("Secret ID rotation successful", func(t *testing.T) {
		mock := &mockVaultRotateSecretIDUtilsBundle{
			t:         t,
			newSecret: "new-secret-id",
			ttl:       time.Hour * 24 * 3, // 3 days
			config: &vaultRotateSecretIdOptions{
				DaysBeforeExpiry: 5,
				SecretStore:      "jenkins",
			},
			updateFuncCalled: false,
		}

		err := runVaultRotateSecretID(mock)
		assert.NoError(t, err)
		assert.True(t, mock.updateFuncCalled)
	})

	t.Run("Secret ID TTL valid, no rotation needed", func(t *testing.T) {
		mock := &mockVaultRotateSecretIDUtilsBundle{
			t:         t,
			newSecret: "new-secret-id",
			ttl:       time.Hour * 24 * 10, // 10 days
			config: &vaultRotateSecretIdOptions{
				DaysBeforeExpiry: 5,
				SecretStore:      "jenkins",
			},
			updateFuncCalled: false,
		}

		err := runVaultRotateSecretID(mock)
		assert.NoError(t, err)
		assert.False(t, mock.updateFuncCalled)
	})

	t.Run("Secret ID expired", func(t *testing.T) {
		mock := &mockVaultRotateSecretIDUtilsBundle{
			t:         t,
			newSecret: "new-secret-id",
			ttl:       0, // expired
			config: &vaultRotateSecretIdOptions{
				DaysBeforeExpiry: 5,
				SecretStore:      "jenkins",
			},
			updateFuncCalled: false,
		}

		err := runVaultRotateSecretID(mock)
		assert.NoError(t, err)
		assert.False(t, mock.updateFuncCalled)
	})

	t.Run("ADO Personal Access Token missing", func(t *testing.T) {
		mock := &mockVaultRotateSecretIDUtilsBundle{
			t:         t,
			newSecret: "new-secret-id",
			ttl:       time.Hour * 24 * 3, // 3 days
			config: &vaultRotateSecretIdOptions{
				DaysBeforeExpiry:       5,
				SecretStore:            "ado",
				AdoPersonalAccessToken: "",
			},
			updateFuncCalled: false,
		}

		err := runVaultRotateSecretID(mock)
		assert.Error(t, err)
		assert.EqualError(t, err, "ADO Personal Access Token is missing")
		assert.False(t, mock.updateFuncCalled)
	})

	t.Run("Error generating new secret ID", func(t *testing.T) {
		mock := &mockVaultRotateSecretIDUtilsBundle{
			t:         t,
			newSecret: "", // Simulate failure to generate new secret ID
			ttl:       time.Hour * 24 * 3, // 3 days
			config: &vaultRotateSecretIdOptions{
				DaysBeforeExpiry: 5,
				SecretStore:      "jenkins",
			},
			updateFuncCalled: false,
		}

		err := runVaultRotateSecretID(mock)
		assert.NoError(t, err)
		assert.False(t, mock.updateFuncCalled)
	})
}

func (v *mockVaultRotateSecretIDUtilsBundle) GenerateNewAppRoleSecret(secretID string, roleName string) (string, error) {
	return v.newSecret, nil
}

func (v *mockVaultRotateSecretIDUtilsBundle) GetAppRoleSecretIDTtl(secretID, roleName string) (time.Duration, error) {
	return v.ttl, nil
}
func (v *mockVaultRotateSecretIDUtilsBundle) GetAppRoleName() (string, error) {
	return "test", nil
}
func (v *mockVaultRotateSecretIDUtilsBundle) UpdateSecretInStore(config *vaultRotateSecretIdOptions, secretID string) error {
	v.updateFuncCalled = true
	assert.Equal(v.t, v.newSecret, secretID)
	return nil
}
func (v *mockVaultRotateSecretIDUtilsBundle) GetConfig() *vaultRotateSecretIdOptions {
	return v.config
}

func getTestConfig() *vaultRotateSecretIdOptions {
	return &vaultRotateSecretIdOptions{
		DaysBeforeExpiry: 5,
	}
}
