//go:build unit

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
