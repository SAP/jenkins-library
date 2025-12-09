//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestVaultIntegration ./integration/...

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/SAP/jenkins-library/pkg/vault"
)

type SecretData = map[string]interface{}

func TestVaultIntegrationGetSecret(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	const testToken = "vault-token"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "vault:1.4.3",
			ExposedPorts:    []string{"8200/tcp"},
			Env:             map[string]string{"VAULT_DEV_ROOT_TOKEN_ID": testToken},
			WaitingFor:      wait.ForLog("Vault server started!").WithStartupTimeout(20 * time.Second)},

		Started: true,
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer vaultContainer.Terminate(ctx)

	ip, err := vaultContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := vaultContainer.MappedPort(ctx, "8200")
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())
	config := &vault.ClientConfig{Config: &api.Config{Address: host}}
	// setup vault for testing
	secretData := SecretData{
		"key1": "value1",
		"key2": "value2",
	}
	setupVault(t, config, testToken, secretData)

	client, err := vault.NewClientWithToken(config, testToken)
	assert.NoError(t, err)
	secret, err := client.GetKvSecret("secret/test")
	assert.NoError(t, err)
	assert.Equal(t, "value1", secret["key1"])
	assert.Equal(t, "value2", secret["key2"])

	secret, err = client.GetKvSecret("kv/test")
	assert.NoError(t, err)
	assert.Equal(t, "value1", secret["key1"])
	assert.Equal(t, "value2", secret["key2"])

}

func TestVaultIntegrationWriteSecret(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	const testToken = "vault-token"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "vault:1.4.3",
			ExposedPorts:    []string{"8200/tcp"},
			Env:             map[string]string{"VAULT_DEV_ROOT_TOKEN_ID": testToken},
			WaitingFor:      wait.ForLog("Vault server started!").WithStartupTimeout(20 * time.Second)},

		Started: true,
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer vaultContainer.Terminate(ctx)

	ip, err := vaultContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := vaultContainer.MappedPort(ctx, "8200")
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())
	config := &vault.ClientConfig{Config: &api.Config{Address: host}}
	// setup vault for testing
	secretData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	client, err := vault.NewClientWithToken(config, testToken)
	assert.NoError(t, err)

	err = client.WriteKvSecret("secret/test", secretData)
	assert.NoError(t, err)

	secret, err := client.GetKvSecret("secret/test")
	assert.NoError(t, err)
	assert.Equal(t, "value1", secret["key1"])
	assert.Equal(t, "value2", secret["key2"])

	// enabling KV engine 1
	vaultClient, err := api.NewClient(config.Config)
	assert.NoError(t, err)
	vaultClient.SetToken(testToken)
	_, err = vaultClient.Logical().Write("sys/mounts/kv", SecretData{
		"path": "kv",
		"type": "kv",
		"options": SecretData{
			"version": "1",
		},
	})
	assert.NoError(t, err)

	err = client.WriteKvSecret("secret/test1", secretData)
	assert.NoError(t, err)

	secret, err = client.GetKvSecret("secret/test1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", secret["key1"])
	assert.Equal(t, "value2", secret["key2"])
}

func TestVaultIntegrationAppRole(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	const testToken = "vault-token"
	const appRolePath = "auth/approle/role/test"
	const appRoleName = "test"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "vault:1.4.3",
			ExposedPorts:    []string{"8200/tcp"},
			Env:             map[string]string{"VAULT_DEV_ROOT_TOKEN_ID": testToken},
			WaitingFor:      wait.ForLog("Vault server started!").WithStartupTimeout(20 * time.Second)},

		Started: true,
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer vaultContainer.Terminate(ctx)

	ip, err := vaultContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := vaultContainer.MappedPort(ctx, "8200")
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())
	config := &vault.ClientConfig{Config: &api.Config{Address: host}}

	secretIDMetadata := map[string]interface{}{
		"field1": "value1",
	}

	roleID, secretID := setupVaultAppRole(t, config, testToken, appRolePath, secretIDMetadata)
	config.RoleID = roleID
	config.SecretID = secretID
	t.Run("Test Vault AppRole login", func(t *testing.T) {
		client, err := vault.NewClient(config)
		assert.NoError(t, err)
		secret, err := client.GetSecret("auth/token/lookup-self")
		meta := secret.Data["meta"].(SecretData)
		assert.Equal(t, meta["field1"], "value1")
		assert.Equal(t, meta["role_name"], "test")
		assert.NoError(t, err)
	})

	t.Run("Test Vault AppRoleTTL Fetch", func(t *testing.T) {
		client, err := vault.NewClientWithToken(config, testToken)
		assert.NoError(t, err)
		ttl, err := client.GetAppRoleSecretIDTtl(secretID, appRoleName)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(90*24*time.Hour), ttl.Round(time.Hour))
	})

	t.Run("Test Vault AppRole Rotation", func(t *testing.T) {
		client, err := vault.NewClientWithToken(config, testToken)
		assert.NoError(t, err)
		newSecretID, err := client.GenerateNewAppRoleSecret(secretID, appRoleName)
		assert.NoError(t, err)
		assert.NotEmpty(t, newSecretID)
		assert.NotEqual(t, secretID, newSecretID)

		// verify metadata is not broken
		client, err = vault.NewClient(config)
		assert.NoError(t, err)
		secret, err := client.GetSecret("auth/token/lookup-self")
		meta := secret.Data["meta"].(SecretData)
		assert.Equal(t, meta["field1"], "value1")
		assert.Equal(t, meta["role_name"], "test")
		assert.NoError(t, err)
	})

	t.Run("Test Fetching RoleName from vault", func(t *testing.T) {
		client, err := vault.NewClient(config)
		assert.NoError(t, err)
		fetchedRoleName, err := client.GetAppRoleName()
		assert.NoError(t, err)
		assert.Equal(t, appRoleName, fetchedRoleName)
	})
}

func TestVaultIntegrationTokenRevocation(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	const testToken = "vault-token"
	const appRolePath = "auth/approle/role/test"
	const appRoleName = "test"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "vault:1.4.3",
			ExposedPorts:    []string{"8200/tcp"},
			Env:             map[string]string{"VAULT_DEV_ROOT_TOKEN_ID": testToken},
			WaitingFor:      wait.ForLog("Vault server started!").WithStartupTimeout(20 * time.Second)},

		Started: true,
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer vaultContainer.Terminate(ctx)

	ip, err := vaultContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := vaultContainer.MappedPort(ctx, "8200")
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())
	config := &vault.ClientConfig{Config: &api.Config{Address: host}}

	secretIDMetadata := map[string]interface{}{
		"field1": "value1",
	}

	roleID, secretID := setupVaultAppRole(t, config, testToken, appRolePath, secretIDMetadata)
	config.RoleID = roleID
	config.SecretID = secretID

	t.Run("Test Revocation works", func(t *testing.T) {
		client, err := vault.NewClient(config)
		assert.NoError(t, err)
		secret, err := client.GetSecret("auth/token/lookup-self")
		meta := secret.Data["meta"].(SecretData)
		assert.Equal(t, meta["field1"], "value1")
		assert.Equal(t, meta["role_name"], "test")
		assert.NoError(t, err)

		err = client.RevokeToken()
		assert.NoError(t, err)

		_, err = client.GetSecret("auth/token/lookup-self")
		assert.Error(t, err)
	})
}

func setupVaultAppRole(t *testing.T, config *vault.ClientConfig, token, appRolePath string, metadata map[string]interface{}) (string, string) {
	t.Helper()
	client, err := api.NewClient(config.Config)
	assert.NoError(t, err)
	client.SetToken(token)
	lClient := client.Logical()

	_, err = lClient.Write("sys/auth/approle", SecretData{
		"type": "approle",
		"config": map[string]interface{}{
			"default_lease_ttl": "7776000s",
			"max_lease_ttl":     "7776000s",
		},
	})
	assert.NoError(t, err)

	_, err = lClient.Write("auth/approle/role/test", SecretData{
		"secret_id_ttl": 7776000,
	})

	assert.NoError(t, err)

	metadataJson, err := json.Marshal(metadata)
	assert.NoError(t, err)

	res, err := lClient.Write("auth/approle/role/test/secret-id", SecretData{
		"metadata": string(metadataJson),
	})

	assert.NoError(t, err)
	secretID := res.Data["secret_id"]

	res, err = lClient.Read("auth/approle/role/test/role-id")
	assert.NoError(t, err)
	roleID := res.Data["role_id"]

	return roleID.(string), secretID.(string)
}

func setupVault(t *testing.T, config *vault.ClientConfig, token string, secret SecretData) {
	t.Helper()
	client, err := api.NewClient(config.Config)
	assert.NoError(t, err)
	client.SetToken(token)

	_, err = client.Logical().Write("secret/data/test", SecretData{"data": secret})
	assert.NoError(t, err)

	// enabling KV engine 1
	_, err = client.Logical().Write("sys/mounts/kv", SecretData{
		"path": "kv",
		"type": "kv",
		"options": SecretData{
			"version": "1",
		},
	})
	assert.NoError(t, err)

	_, err = client.Logical().Write("kv/test", secret)
	assert.NoError(t, err)

}
