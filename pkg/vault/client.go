package vault

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
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
func NewClient(config *api.Config, token, namespace string) (Client, error) {
	if config == nil {
		config = api.DefaultConfig()
	}
	client, err := api.NewClient(config)
	if err != nil {
		return Client{}, err
	}

	if namespace != "" {
		client.SetNamespace(namespace)
	}

	client.SetToken(token)
	return Client{client.Logical()}, nil
}

// NewClientWithAppRole instantiates a new client and obtains a token via the AppRole auth method
func NewClientWithAppRole(config *api.Config, roleID, secretID, namespace string) (Client, error) {
	if config == nil {
		config = api.DefaultConfig()
	}

	client, err := api.NewClient(config)
	if err != nil {
		return Client{}, err
	}

	if namespace != "" {
		client.SetNamespace(namespace)
	}

	log.Entry().Debug("Using approle login")
	result, err := client.Logical().Write("auth/approle/login", map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	})

	if err != nil {
		return Client{}, err
	}

	authInfo := result.Auth
	if authInfo == nil || authInfo.ClientToken == "" {
		return Client{}, fmt.Errorf("Could not obtain token from approle with role_id %s", roleID)
	}

	log.Entry().Debugf("Login to vault %s in namespace %s successfull", config.Address, namespace)
	return NewClient(config, authInfo.ClientToken, namespace)
}

// GetSecret uses the given path to fetch a secret from vault
func (v Client) GetSecret(path string) (*api.Secret, error) {
	path = sanitizePath(path)
	c := v.lClient

	secret, err := c.Read(path)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// GetKvSecret reads secret from the KV engine.
// It Automatically transforms the logical path to the HTTP API Path for the corresponding KV Engine version
func (v Client) GetKvSecret(path string) (map[string]string, error) {
	path = sanitizePath(path)
	mountpath, version, err := v.getKvInfo(path)
	if err != nil {
		return nil, err
	}
	if version == 2 {
		path = addPrefixToKvPath(path, mountpath, "data")
	} else if version != 1 {
		return nil, fmt.Errorf("KV Engine in version %d is currently not supported", version)
	}

	secret, err := v.GetSecret(path)
	if secret == nil || err != nil {
		return nil, err

	}
	var rawData interface{}
	switch version {
	case 1:
		rawData = secret.Data
	case 2:
		var ok bool
		rawData, ok = secret.Data["data"]
		if !ok {
			return nil, fmt.Errorf("Missing 'data' field in response: %v", rawData)
		}
	}

	data, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Excpected 'data' field to be a map[string]interface{} but got %T instead", rawData)
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

func addPrefixToKvPath(p, mountPath, apiPrefix string) string {
	switch {
	case p == mountPath, p == strings.TrimSuffix(mountPath, "/"):
		return path.Join(mountPath, apiPrefix)
	default:
		p = strings.TrimPrefix(p, mountPath)
		return path.Join(mountPath, apiPrefix, p)
	}
}

func (v *Client) getKvInfo(path string) (string, int, error) {
	secret, err := v.GetSecret("sys/internal/ui/mounts/" + path)
	if err != nil {
		return "", 0, err
	}

	if secret == nil {
		return "", 0, fmt.Errorf("Failed to get version and engine mountpoint for path: %s", path)
	}

	var mountPath string
	if mountPathRaw, ok := secret.Data["path"]; ok {
		mountPath = mountPathRaw.(string)
	}

	options := secret.Data["options"]
	if options == nil {
		return mountPath, 1, nil
	}

	versionRaw := options.(map[string]interface{})["version"]
	if versionRaw == nil {
		return mountPath, 1, nil
	}

	version := versionRaw.(string)
	if version == "" {
		return mountPath, 1, nil
	}

	vNumber, err := strconv.Atoi(version)
	if err != nil {
		return mountPath, 0, err
	}

	return mountPath, vNumber, nil
}

func sanitizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return path
}
