//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockVaultRotateSecretIDUtilsBundle struct {
	t                *testing.T
	newSecret        string
	ttl              time.Duration
	config           *vaultRotateSecretIdOptions
	updateFuncCalled bool
	UpdateSecretFunc func(config *vaultRotateSecretIdOptions, secretID string) error // Function field
}

// func TestRunVaultRotateSecretId(t *testing.T) {
// 	t.Parallel()
// 	mock := &mockVaultRotateSecretIDUtilsBundle{t, "test-secret", time.Hour, getTestConfig(), false}
// 	runVaultRotateSecretID(mock)
// 	assert.True(t, mock.updateFuncCalled)
// }

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
		assert.True(t, mock.updateFuncCalled)
	})

	t.Run("ADO Personal Access Token missing and automaticd didn't updated secrets", func(t *testing.T) {
		mock := &mockVaultRotateSecretIDUtilsBundle{
			t:         t,
			newSecret: "new-secret-id",
			ttl:       time.Hour * 24 * 16, // 16 days
			config: &vaultRotateSecretIdOptions{
				DaysBeforeExpiry:       15,
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

	t.Run("Error updating secret in store", func(t *testing.T) {
		mock := &mockVaultRotateSecretIDUtilsBundle{
			t:         t,
			newSecret: "new-secret-id",
			ttl:       time.Hour * 24 * 3, // 3 days
			config: &vaultRotateSecretIdOptions{
				DaysBeforeExpiry: 15,
				SecretStore:      "jenkins",
			},
			updateFuncCalled: false,
			UpdateSecretFunc: func(config *vaultRotateSecretIdOptions, secretID string) error {
				return fmt.Errorf("failed to update secret in store")
			}, // Override the behavior
		}

		err := runVaultRotateSecretID(mock)
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to update secret in store")
		assert.True(t, mock.updateFuncCalled)
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

	// Call the overridden function if it is set
	if v.UpdateSecretFunc != nil {
		return v.UpdateSecretFunc(config, secretID)
	}

	// Default behavior
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
