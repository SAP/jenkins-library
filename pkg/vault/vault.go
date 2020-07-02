package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Client handles communication with Vault
type Client struct {
	lClient logicalClient
}

// logicalClient interface for mocking
type logicalClient interface {
	Read(string) (*api.Secret, error)
}

// NewClient instantiates a Client and sets the specified token
func NewClient(config *api.Config, token string) (Client, error) {
	if config == nil {
		config = api.DefaultConfig()
	}
	client, err := api.NewClient(config)
	if err != nil {
		return Client{}, err
	}

	client.SetToken(token)
	return Client{client.Logical()}, nil
}

// GetSecret uses the given path to fetch a secret from vault
func (v Client) GetSecret(path string) (*api.Secret, error) {
	c := v.lClient

	secret, err := c.Read(path)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// GetKVSecret reads a secret from the vault KV engine (v2) and retruns the KV pairs
func (v Client) GetKVSecret(path string) (map[string]string, error) {
	secret, err := v.GetSecret(path)
	if err != nil {
		return nil, err
	}
	rawData, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("Missing 'data' field in response: %v", rawData)
	}
	data, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Excpected 'data' field to be a map but got %T instead", data)
	}

	secretData := make(map[string]string, len(data))
	for k, v := range data {
		valueStr, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("Expected secret value to be a string but got %T instead", v)
		}
		secretData[k] = valueStr
	}

	return secretData, nil

}
