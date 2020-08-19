// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type SecretData = map[string]interface{}

func TestGetVaultSecret(t *testing.T) {
	t.Parallel()
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
	config := &api.Config{Address: host}
	// setup vault for testing
	secretData := SecretData{
		"key1": "value1",
		"key2": "value2",
	}
	setupVault(t, config, testToken, secretData)

	client, err := vault.NewClient(config, testToken, "")
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

func setupVault(t *testing.T, config *api.Config, token string, secret SecretData) {
	t.Helper()
	client, err := api.NewClient(config)
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
