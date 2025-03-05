package vault

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/hashicorp/vault/api"
	"path"
	"strconv"
	"strings"
	"time"
)

// GetSecret uses the given path to fetch a secret from vault
func (c *Client) GetSecret(path string) (*api.Secret, error) {
	return c.logical.Read(sanitizePath(path))
}

// GetKvSecret reads secret from the KV engine.
// It Automatically transforms the logical path to the HTTP API Path for the corresponding KV Engine version
func (c *Client) GetKvSecret(path string) (map[string]string, error) {
	path = sanitizePath(path)
	mountPath, version, err := c.getKvInfo(path)
	if err != nil {
		return nil, err
	}
	if version == 2 {
		path = addPrefixToKvPath(path, mountPath, "data")
	} else if version != 1 {
		return nil, fmt.Errorf("KV Engine in version %d is currently not supported", version)
	}

	secret, err := c.GetSecret(path)
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
			return nil, fmt.Errorf("missing 'data' field in response: %v", rawData)
		}
	}

	data, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("excpected 'data' field to be a map[string]interface{} but got %T instead", rawData)
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
func (c *Client) WriteKvSecret(path string, newSecret map[string]string) error {
	oldSecret, err := c.GetKvSecret(path)
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
	mountPath, version, err := c.getKvInfo(path)
	if err != nil {
		return err
	}
	if version == 2 {
		path = addPrefixToKvPath(path, mountPath, "data")
		secret = map[string]interface{}{"data": secret}
	} else if version != 1 {
		return fmt.Errorf("KV Engine in version %d is currently not supported", version)
	}

	_, err = c.logical.Write(path, secret)
	return err
}

// GenerateNewAppRoleSecret creates a new secret-id
func (c *Client) GenerateNewAppRoleSecret(secretID, appRoleName string) (string, error) {
	appRolePath := c.getAppRolePath(appRoleName)
	secretIDData, err := c.lookupSecretID(secretID, appRolePath)
	if err != nil {
		return "", err
	}

	reqPath := sanitizePath(path.Join(appRolePath, "/secret-id"))

	// we preserve metadata which was attached to the secret-id
	jsonBytes, err := json.Marshal(secretIDData["metadata"])
	if err != nil {
		return "", err
	}

	secret, err := c.logical.Write(reqPath, map[string]interface{}{
		"metadata": string(jsonBytes),
	})
	if err != nil {
		return "", err
	}

	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("could not generate new approle secret-id for approle path %s", reqPath)
	}

	secretIDRaw, ok := secret.Data["secret_id"]
	if !ok {
		return "", fmt.Errorf("Vault response for path %s did not contain a new secret-id", reqPath)
	}

	newSecretID, ok := secretIDRaw.(string)
	if !ok {
		return "", fmt.Errorf("new secret-id from approle path %s has an unexpected type %T expected 'string'", reqPath, secretIDRaw)
	}

	// TODO: remove after testing
	log.Entry().Debugf("GenerateNewAppRoleSecret - secretID: %#v", secretID)
	log.Entry().Debugf("GenerateNewAppRoleSecret - appRoleName: %#v", appRoleName)
	log.Entry().Debugf("GenerateNewAppRoleSecret - appRolePath: %#v", appRolePath)
	log.Entry().Debugf("GenerateNewAppRoleSecret - secretIDData: %#v", secretIDData)
	log.Entry().Debugf("GenerateNewAppRoleSecret - new secret data: %#v", secret)
	log.Entry().Debugf("GenerateNewAppRoleSecret - new secret ID: %#v", newSecretID)

	return newSecretID, nil
}

// GetAppRoleSecretIDTtl returns the remaining time until the given secret-id expires
func (c *Client) GetAppRoleSecretIDTtl(secretID, roleName string) (time.Duration, error) {
	appRolePath := c.getAppRolePath(roleName)
	data, err := c.lookupSecretID(secretID, appRolePath)
	if err != nil {
		return 0, err
	}

	if data == nil || data["expiration_time"] == nil {
		return 0, fmt.Errorf("could not load secret-id information from path %s", appRolePath)
	}

	expiration, ok := data["expiration_time"].(string)
	if !ok || expiration == "" {
		return 0, fmt.Errorf("could not handle get expiration time for secret-id from path %s", appRolePath)
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
func (c *Client) RevokeToken() error {
	_, err := c.logical.Write("auth/token/revoke-self", map[string]interface{}{})
	return err
}

// MustRevokeToken same as RevokeToken but the program is terminated with an error if this fails.
// Should be used in defer statements only.
func (c *Client) MustRevokeToken() {
	lookupPath := "auth/token/lookup-self"
	const serviceTokenPrefix = "hvs."

	secret, err := c.GetSecret(lookupPath)
	if err != nil {
		log.Entry().Warnf("Could not lookup token at %s, not continuing to revoke: %v", lookupPath, err)
		return
	}

	tokenID, ok := secret.Data["id"].(string)
	if !ok {
		log.Entry().Warnf("Could not lookup token.Data.id at %s, not continuing to revoke", lookupPath)
		return
	}

	if !strings.HasPrefix(tokenID, serviceTokenPrefix) {
		log.Entry().Warnf("Service token not identified at %s, not continuing to revoke", lookupPath)
		return
	}

	if err = c.RevokeToken(); err != nil {
		log.Entry().WithError(err).Fatal("Could not revoke token")
	}
}

// GetAppRoleName returns the AppRole role name which was used to authenticate.
// Returns "" when AppRole authentication wasn't used
func (c *Client) GetAppRoleName() (string, error) {
	const lookupPath = "auth/token/lookup-self"
	secret, err := c.GetSecret(lookupPath)
	if err != nil {
		return "", err
	}

	if secret.Data == nil {
		return "", fmt.Errorf("could not lookup token information: %s", lookupPath)
	}

	meta, ok := secret.Data["meta"]

	if !ok {
		return "", fmt.Errorf("token info did not contain metadata %s", lookupPath)
	}

	metaMap, ok := meta.(map[string]interface{})

	if !ok {
		return "", fmt.Errorf("token info field 'meta' is not a map: %s", lookupPath)
	}

	roleName := metaMap["role_name"]

	if roleName == nil {
		return "", nil
	}

	roleNameStr, ok := roleName.(string)
	if !ok {
		// when AppRole authentication is not used vault admins can use the role_name field with other type
		// so no error in this case
		return "", nil
	}

	return roleNameStr, nil
}

func (c *Client) getAppRolePath(roleName string) string {
	appRoleMountPoint := c.cfg.AppRoleMountPoint
	if appRoleMountPoint == "" {
		appRoleMountPoint = "auth/approle"
	}
	return path.Join(appRoleMountPoint, "role", roleName)
}

func (c *Client) getKvInfo(path string) (string, int, error) {
	secret, err := c.GetSecret("sys/internal/ui/mounts/" + path)
	if err != nil {
		return "", 0, err
	}

	if secret == nil {
		return "", 0, fmt.Errorf("failed to get version and engine mountpoint for path: %s", path)
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

func (c *Client) lookupSecretID(secretID, appRolePath string) (map[string]interface{}, error) {
	reqPath := sanitizePath(path.Join(appRolePath, "/secret-id/lookup"))
	secret, err := c.logical.Write(reqPath, map[string]interface{}{
		"secret_id": secretID,
	})
	if err != nil {
		return nil, err
	}

	return secret.Data, nil
}
