package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/hashicorp/vault/api"
)

// Client handles communication with Vault
type Client struct {
	lClient logicalClient
	config  *Config
}

// Config contains the vault client configuration
type Config struct {
	*api.Config
	AppRoleMountPoint string
	Namespace         string
}

// logicalClient interface for mocking
type logicalClient interface {
	Read(string) (*api.Secret, error)
	Write(string, map[string]interface{}) (*api.Secret, error)
}

type VaultCredentials struct {
	AppRoleID       string
	AppRoleSecretID string
	VaultToken      string
}

// NewClient instantiates a Client and sets the specified token
func NewClient(config *Config, token string) (Client, error) {
	if config == nil {
		config = &Config{Config: api.DefaultConfig()}
	}
	client, err := api.NewClient(config.Config)
	if err != nil {
		return Client{}, err
	}
	if config.Namespace != "" {
		client.SetNamespace(config.Namespace)
	}
	client.SetToken(token)
	return Client{client.Logical(), config}, nil
}

// NewClientWithAppRole instantiates a new client and obtains a token via the AppRole auth method
func NewClientWithAppRole(config *Config, roleID, secretID string) (Client, error) {
	if config == nil {
		config = &Config{Config: api.DefaultConfig()}
	}
	if config.AppRoleMountPoint == "" {
		config.AppRoleMountPoint = "auth/approle"
	}
	client, err := api.NewClient(config.Config)
	if err != nil {
		return Client{}, err
	}

	client.SetMinRetryWait(time.Second * 5)
	client.SetMaxRetryWait(time.Second * 90)
	client.SetMaxRetries(3)
	client.SetCheckRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil {
			log.Entry().Debugln("Vault response: ", resp.Status, resp.StatusCode, err)
		} else {
			log.Entry().Debugln("Vault response: ", err)
		}

		isEOF := false
		if err != nil && strings.Contains(err.Error(), "EOF") {
			log.Entry().Infoln("isEOF is true")
			isEOF = true
		}

		if err == io.EOF {
			log.Entry().Infoln("err = io.EOF is true")
		}

		retry, err := api.DefaultRetryPolicy(ctx, resp, err)

		if err != nil || err == io.EOF || isEOF || retry {
			log.Entry().Infoln("Retrying vault request...")
			return true, nil
		}
		return false, nil
	})

	if config.Namespace != "" {
		client.SetNamespace(config.Namespace)
	}

	result, err := client.Logical().Write(path.Join(config.AppRoleMountPoint, "/login"), map[string]interface{}{
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

	return NewClient(config, authInfo.ClientToken)
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
		switch t := v.(type) {
		case string:
			secretData[k] = t
		case int:
			secretData[k] = fmt.Sprintf("%d", t)
		default:
			jsonBytes, err := json.Marshal(t)
			if err != nil {
				log.Entry().Warnf("failed to parse Vault secret key %q, error: %s", k, err.Error())
				continue
			}

			secretData[k] = string(jsonBytes)
		}
	}
	return secretData, nil
}

// WriteKvSecret writes secret to kv engine
func (v Client) WriteKvSecret(path string, newSecret map[string]string) error {
	oldSecret, err := v.GetKvSecret(path)
	if err != nil {
		return err
	}
	secret := make(map[string]interface{}, len(oldSecret))
	for k, v := range oldSecret {
		secret[k] = v
	}
	for k, v := range newSecret {
		secret[k] = v
	}
	path = sanitizePath(path)
	mountpath, version, err := v.getKvInfo(path)
	if err != nil {
		return err
	}
	if version == 2 {
		path = addPrefixToKvPath(path, mountpath, "data")
		secret = map[string]interface{}{"data": secret}
	} else if version != 1 {
		return fmt.Errorf("KV Engine in version %d is currently not supported", version)
	}

	_, err = v.lClient.Write(path, secret)
	return err
}

// GenerateNewAppRoleSecret creates a new secret-id
func (v *Client) GenerateNewAppRoleSecret(secretID, appRoleName string) (string, error) {
	appRolePath := v.getAppRolePath(appRoleName)
	secretIDData, err := v.lookupSecretID(secretID, appRolePath)
	if err != nil {
		return "", err
	}

	reqPath := sanitizePath(path.Join(appRolePath, "/secret-id"))

	// we preserve metadata which was attached to the secret-id
	json, err := json.Marshal(secretIDData["metadata"])
	if err != nil {
		return "", err
	}
	secret, err := v.lClient.Write(reqPath, map[string]interface{}{
		"metadata": string(json),
	})

	if err != nil {
		return "", err
	}

	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("Could not generate new approle secret-id for approle path %s", reqPath)
	}

	secretIDRaw, ok := secret.Data["secret_id"]
	if !ok {
		return "", fmt.Errorf("Vault response for path %s did not contain a new secret-id", reqPath)
	}

	newSecretID, ok := secretIDRaw.(string)
	if !ok {
		return "", fmt.Errorf("New secret-id from approle path %s has an unexpected type %T expected 'string'", reqPath, secretIDRaw)
	}

	return newSecretID, nil
}

// GetAppRoleSecretIDTtl returns the remaining time until the given secret-id expires
func (v *Client) GetAppRoleSecretIDTtl(secretID, roleName string) (time.Duration, error) {
	appRolePath := v.getAppRolePath(roleName)
	data, err := v.lookupSecretID(secretID, appRolePath)
	if err != nil {
		return 0, err
	}

	if data == nil || data["expiration_time"] == nil {
		return 0, fmt.Errorf("Could not load secret-id information from path %s", appRolePath)
	}

	expiration, ok := data["expiration_time"].(string)
	if !ok || expiration == "" {
		return 0, fmt.Errorf("Could not handle get expiration time for secret-id from path %s", appRolePath)
	}

	expirationDate, err := time.Parse(time.RFC3339, expiration)

	if err != nil {
		return 0, err
	}

	ttl := expirationDate.Sub(time.Now())
	if ttl < 0 {
		return 0, nil
	}

	return ttl, nil
}

// RevokeToken revokes the token which is currently used.
// The client can't be used anymore after this function was called.
func (v Client) RevokeToken() error {
	_, err := v.lClient.Write("auth/token/revoke-self", map[string]interface{}{})
	return err
}

// MustRevokeToken same as RevokeToken but the programm is terminated with an error if this fails.
// Should be used in defer statements only.
func (v Client) MustRevokeToken() {
	if err := v.RevokeToken(); err != nil {
		log.Entry().WithError(err).Fatal("Could not revoke token")
	}
}

// GetAppRoleName returns the AppRole role name which was used to authenticate.
// Returns "" when AppRole authentication wasn't used
func (v *Client) GetAppRoleName() (string, error) {
	const lookupPath = "auth/token/lookup-self"
	secret, err := v.GetSecret(lookupPath)
	if err != nil {
		return "", err
	}

	if secret.Data == nil {
		return "", fmt.Errorf("Could not lookup token information: %s", lookupPath)
	}

	meta, ok := secret.Data["meta"]

	if !ok {
		return "", fmt.Errorf("Token info did not contain metadata %s", lookupPath)
	}

	metaMap, ok := meta.(map[string]interface{})

	if !ok {
		return "", fmt.Errorf("Token info field 'meta' is not a map: %s", lookupPath)
	}

	roleName := metaMap["role_name"]

	if roleName == nil {
		return "", nil
	}

	roleNameStr, ok := roleName.(string)
	if !ok {
		// when approle authentication is not used vault admins can use the role_name field with other type
		// so no error in this case
		return "", nil
	}

	return roleNameStr, nil
}

// SetAppRoleMountPoint sets the path under which the approle auth backend is mounted
func (v *Client) SetAppRoleMountPoint(appRoleMountpoint string) {
	v.config.AppRoleMountPoint = appRoleMountpoint
}

func (v *Client) getAppRolePath(roleName string) string {
	appRoleMountPoint := v.config.AppRoleMountPoint
	if appRoleMountPoint == "" {
		appRoleMountPoint = "auth/approle"
	}
	return path.Join(appRoleMountPoint, "role", roleName)
}

func sanitizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return path
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

func (v *Client) lookupSecretID(secretID, appRolePath string) (map[string]interface{}, error) {
	reqPath := sanitizePath(path.Join(appRolePath, "/secret-id/lookup"))
	secret, err := v.lClient.Write(reqPath, map[string]interface{}{
		"secret_id": secretID,
	})
	if err != nil {
		return nil, err
	}

	return secret.Data, nil
}
