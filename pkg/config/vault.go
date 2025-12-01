package config

import (
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config/interpolation"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"
)

const (
	vaultRootPaths                      = "vaultRootPaths"
	vaultTestCredentialPath             = "vaultTestCredentialPath"
	vaultCredentialPath                 = "vaultCredentialPath"
	vaultTestCredentialKeys             = "vaultTestCredentialKeys"
	vaultCredentialKeys                 = "vaultCredentialKeys"
	vaultAppRoleID                      = "vaultAppRoleID"
	vaultAppRoleSecretID                = "vaultAppRoleSecreId"
	vaultServerUrl                      = "vaultServerUrl"
	vaultNamespace                      = "vaultNamespace"
	vaultBasePath                       = "vaultBasePath"
	vaultPipelineName                   = "vaultPipelineName"
	vaultPath                           = "vaultPath"
	skipVault                           = "skipVault"
	vaultDisableOverwrite               = "vaultDisableOverwrite"
	vaultTestCredentialEnvPrefix        = "vaultTestCredentialEnvPrefix"
	vaultCredentialEnvPrefix            = "vaultCredentialEnvPrefix"
	vaultTestCredentialEnvPrefixDefault = "PIPER_TESTCREDENTIAL_"
	VaultCredentialEnvPrefixDefault     = "PIPER_VAULTCREDENTIAL_"
	vaultSecretName                     = ".+VaultSecretName$"
)

var (
	vaultFilter = []string{
		vaultRootPaths,
		vaultAppRoleID,
		vaultAppRoleSecretID,
		vaultServerUrl,
		vaultNamespace,
		vaultBasePath,
		vaultPipelineName,
		vaultPath,
		skipVault,
		vaultDisableOverwrite,
		vaultTestCredentialPath,
		vaultTestCredentialKeys,
		vaultTestCredentialEnvPrefix,
		vaultCredentialPath,
		vaultCredentialKeys,
		vaultCredentialEnvPrefix,
		vaultSecretName,
	}

	// VaultRootPaths are the lookup paths piper tries to use during the vault lookup.
	// A path is only used if it's variables can be interpolated from the config
	VaultRootPaths = []string{
		"$(vaultPath)",
		"$(vaultBasePath)/$(vaultPipelineName)",
		"$(vaultBasePath)/GROUP-SECRETS",
	}

	// VaultSecretFileDirectory holds the directory for the current step run to temporarily store secret files fetched from vault
	VaultSecretFileDirectory = ""
)

// VaultCredentials hold all the auth information needed to fetch configuration from vault
type VaultCredentials struct {
	AppRoleID       string
	AppRoleSecretID string
	VaultToken      string
}

// VaultClient interface for mocking
type VaultClient interface {
	GetKvSecret(string) (map[string]string, error)
	MustRevokeToken()
	GetOIDCTokenByValidation(string) (string, error)
}

// globalVaultClient is supposed to be used in the steps code.
var globalVaultClient *vault.Client

func GlobalVaultClient() VaultClient {
	// an interface containing a nil pointer is considered non-nil in Go
	// It is nil if Vault is not configured
	if globalVaultClient == nil {
		return nil
	}

	return globalVaultClient
}

func (s *StepConfig) mixinVaultConfig(parameters []StepParameters, configs ...map[string]interface{}) {
	for _, config := range configs {
		s.mixIn(config, vaultFilter, StepData{})
		// when an empty filter is returned we skip the mixin call since an empty filter will allow everything
		if referencesFilter := getFilterForResourceReferences(parameters); len(referencesFilter) > 0 {
			s.mixIn(config, referencesFilter, StepData{})
		}
	}
}

// GetVaultClientFromConfig logs in to Vault and returns authorized Vault client.
// It's important to revoke token provided to this client after usage.
// Currently, revocation will happen at the end of each step execution (see _generated.go part of the steps)
func GetVaultClientFromConfig(config map[string]interface{}, creds VaultCredentials) (VaultClient, error) {
	address, addressOk := config["vaultServerUrl"].(string)
	// if vault isn't used it's not an error
	if !addressOk || creds.VaultToken == "" && (creds.AppRoleID == "" || creds.AppRoleSecretID == "") {
		log.Entry().Debug("Vault not configured")
		return nil, nil
	}
	log.Entry().WithField("vaultURL", address).Info("Logging into Vault")
	namespace := ""
	// namespaces are only available in vault enterprise so using them should be optional
	if config["vaultNamespace"] != nil {
		namespace = config["vaultNamespace"].(string)
		log.Entry().WithField("namespace", namespace).Debug("Vault namespace configured")
	}
	var client *vault.Client
	var err error
	clientConfig := &vault.ClientConfig{Config: &api.Config{Address: address}, Namespace: namespace}
	if creds.VaultToken != "" {
		log.Entry().Debug("Using Vault token authentication")
		client, err = vault.NewClientWithToken(clientConfig, creds.VaultToken)
	} else {
		log.Entry().Debug("Using Vault AppRole authentication")
		clientConfig.RoleID = creds.AppRoleID
		clientConfig.SecretID = creds.AppRoleSecretID
		client, err = vault.NewClient(clientConfig)
	}
	if err != nil {
		log.Entry().WithError(err).Error("Vault authentication failed")
		return nil, err
	}

	// Set global vault client for usage in steps
	globalVaultClient = client

	log.Entry().Info("Vault authentication succeeded")
	return client, nil
}

func resolveAllVaultReferences(config *StepConfig, client VaultClient, params []StepParameters) {
	for _, param := range params {
		if ref := param.GetReference("vaultSecret"); ref != nil {
			resolveVaultReference(ref, config, client, param)
		}
		if ref := param.GetReference("vaultSecretFile"); ref != nil {
			resolveVaultReference(ref, config, client, param)
		}
	}
}

func resolveVaultReference(ref *ResourceReference, config *StepConfig, client VaultClient, param StepParameters) {
	vaultDisableOverwrite, _ := config.Config["vaultDisableOverwrite"].(bool)
	if paramValue, _ := config.Config[param.Name].(string); vaultDisableOverwrite && paramValue != "" {
		log.Entry().Debugf("Not fetching '%s' from Vault since it has already been set", param.Name)
		return
	}

	log.Entry().WithField("parameter", param.Name).Debug("Resolving vault secret")

	var secretValue *string
	for _, vaultPath := range getSecretReferencePaths(ref, config.Config) {
		// it should be possible to configure the root path were the secret is stored
		vaultPath, ok := interpolation.ResolveString(vaultPath, config.Config)
		if !ok {
			continue
		}

		secretValue = lookupPath(client, vaultPath, &param)
		if secretValue != nil {
			log.Entry().WithField("vaultPath", vaultPath).Debug("Vault secret resolved successfully")
			switch ref.Type {
			case "vaultSecret":
				config.Config[param.Name] = *secretValue
			case "vaultSecretFile":
				filePath, err := createTemporarySecretFile(param.Name, *secretValue)
				if err != nil {
					log.Entry().WithError(err).Warnf("Couldn't create temporary secret file for '%s'", param.Name)
					return
				}
				config.Config[param.Name] = filePath
			}
			break
		}
	}
	if secretValue == nil {
		log.Entry().WithField("parameter", param.Name).Info("Secret not found in Vault at configured paths.")
	}
}

func resolveVaultTestCredentialsWrapper(config *StepConfig, client VaultClient) {
	log.Entry().Debug("Resolving test credentials from Vault")
	resolveVaultCredentialsWrapperBase(config, client, vaultTestCredentialPath, vaultTestCredentialKeys, vaultTestCredentialEnvPrefix, resolveVaultTestCredentials)
}

func resolveVaultCredentialsWrapper(config *StepConfig, client VaultClient) {
	log.Entry().Debug("Resolving credentials from Vault")
	resolveVaultCredentialsWrapperBase(config, client, vaultCredentialPath, vaultCredentialKeys, vaultCredentialEnvPrefix, resolveVaultCredentials)
}

func resolveVaultCredentialsWrapperBase(
	config *StepConfig, client VaultClient,
	vaultCredPath, vaultCredKeys, vaultCredEnvPrefix string,
	resolveVaultCredentials func(config *StepConfig, client VaultClient),
) {
	switch config.Config[vaultCredPath].(type) {
	case string:
		resolveVaultCredentials(config, client)
	case []interface{}:
		vaultCredentialPathCopy := config.Config[vaultCredPath].([]interface{})
		vaultCredentialKeysCopy, keysOk := config.Config[vaultCredKeys].([]interface{})
		vaultCredentialEnvPrefixCopy, prefixOk := config.Config[vaultCredEnvPrefix].([]interface{})

		if !keysOk {
			log.Entry().Debug("Vault credential resolution failed: unknown type of keys")
			return
		}

		if len(vaultCredentialKeysCopy) != len(vaultCredentialPathCopy) {
			log.Entry().Debug("Vault credential resolution failed: mismatched count of values and keys")
			return
		}

		if prefixOk && len(vaultCredentialEnvPrefixCopy) != len(vaultCredentialPathCopy) {
			log.Entry().Debug("Vault credential resolution failed: mismatched count of values and environment prefixes")
			return
		}

		for i := 0; i < len(vaultCredentialPathCopy); i++ {
			if prefixOk {
				config.Config[vaultCredEnvPrefix] = vaultCredentialEnvPrefixCopy[i]
			}
			config.Config[vaultCredPath] = vaultCredentialPathCopy[i]
			config.Config[vaultCredKeys] = vaultCredentialKeysCopy[i]
			resolveVaultCredentials(config, client)
		}

		config.Config[vaultCredPath] = vaultCredentialPathCopy
		config.Config[vaultCredKeys] = vaultCredentialKeysCopy
		config.Config[vaultCredEnvPrefix] = vaultCredentialEnvPrefixCopy
	default:
		log.Entry().Debug("Vault credential resolution failed: unknown type of path")
		return
	}
}

// resolve test credential keys and expose as environment variables
func resolveVaultTestCredentials(config *StepConfig, client VaultClient) {
	credPath, pathOk := config.Config[vaultTestCredentialPath].(string)
	keys := getTestCredentialKeys(config)
	if !(pathOk && keys != nil) || credPath == "" || len(keys) == 0 {
		log.Entry().Debugf("Not fetching test credentials from Vault since they are not (properly) configured")
		return
	}

	lookupPath := make([]string, 3)
	lookupPath[0] = "$(vaultPath)/" + credPath
	lookupPath[1] = "$(vaultBasePath)/$(vaultPipelineName)/" + credPath
	lookupPath[2] = "$(vaultBasePath)/GROUP-SECRETS/" + credPath

	for _, path := range lookupPath {
		vaultPath, ok := interpolation.ResolveString(path, config.Config)
		if !ok {
			continue
		}

		secret, err := client.GetKvSecret(vaultPath)
		if err != nil {
			log.Entry().WithError(err).WithField("vaultPath", vaultPath).Debug("Failed to fetch secret from vault path")
			continue
		}
		if secret == nil {
			continue
		}
		secretsResolved := false
		secretsResolved = populateTestCredentialsAsEnvs(config, secret, keys)
		if secretsResolved {
			// prevent overwriting resolved secrets
			// only allows vault test credentials on one / the same vault path
			break
		}
	}
}

func resolveVaultCredentials(config *StepConfig, client VaultClient) {
	credPath, pathOk := config.Config[vaultCredentialPath].(string)
	keys := getCredentialKeys(config)
	if !(pathOk && keys != nil) || credPath == "" || len(keys) == 0 {
		log.Entry().Debugf("Not fetching credentials from vault since they are not (properly) configured")
		return
	}

	lookupPath := make([]string, 3)
	lookupPath[0] = "$(vaultPath)/" + credPath
	lookupPath[1] = "$(vaultBasePath)/$(vaultPipelineName)/" + credPath
	lookupPath[2] = "$(vaultBasePath)/GROUP-SECRETS/" + credPath

	for _, path := range lookupPath {
		vaultPath, ok := interpolation.ResolveString(path, config.Config)
		if !ok {
			continue
		}

		secret, err := client.GetKvSecret(vaultPath)
		if err != nil {
			log.Entry().WithError(err).Debugf("Couldn't fetch secret at '%s'", vaultPath)
			continue
		}
		if secret == nil {
			continue
		}
		secretsResolved := false
		secretsResolved = populateCredentialsAsEnvs(config, secret, keys)
		if secretsResolved {
			// prevent overwriting resolved secrets
			// only allows vault test credentials on one / the same vault path
			break
		}
	}
}

func populateTestCredentialsAsEnvs(config *StepConfig, secret map[string]string, keys []string) (matched bool) {
	vaultTestCredentialEnvPrefix, ok := config.Config["vaultTestCredentialEnvPrefix"].(string)
	if !ok || len(vaultTestCredentialEnvPrefix) == 0 {
		vaultTestCredentialEnvPrefix = vaultTestCredentialEnvPrefixDefault
	}
	for secretKey, secretValue := range secret {
		for _, key := range keys {
			if secretKey == key {
				log.RegisterSecret(secretValue)
				envVariable := vaultTestCredentialEnvPrefix + ConvertEnvVar(secretKey)
				log.Entry().Debugf("Exposing test credential '%v' as '%v'", key, envVariable)
				os.Setenv(envVariable, secretValue)
				matched = true
			}
		}
	}
	return
}

func populateCredentialsAsEnvs(config *StepConfig, secret map[string]string, keys []string) (matched bool) {
	vaultCredentialEnvPrefix, ok := config.Config["vaultCredentialEnvPrefix"].(string)
	isCredentialEnvPrefixDefault := false

	if !ok {
		vaultCredentialEnvPrefix = VaultCredentialEnvPrefixDefault
		isCredentialEnvPrefixDefault = true
	}
	for secretKey, secretValue := range secret {
		for _, key := range keys {
			if secretKey == key {
				log.RegisterSecret(secretValue)
				envVariable := vaultCredentialEnvPrefix + ConvertEnvVar(secretKey)
				log.Entry().Debugf("Exposing general purpose credential '%v' as '%v'", key, envVariable)
				os.Setenv(envVariable, secretValue)

				log.RegisterSecret(piperutils.EncodeString(secretValue))
				envVariable = vaultCredentialEnvPrefix + ConvertEnvVar(secretKey) + "_BASE64"
				log.Entry().Debugf("Exposing general purpose base64 encoded credential '%v' as '%v'", key, envVariable)
				os.Setenv(envVariable, piperutils.EncodeString(secretValue))
				matched = true
			}
		}
	}

	// we always create a standard env variable with the default prefx so that
	// we can always refer to it in steps if its to be hard-coded
	if !isCredentialEnvPrefixDefault {
		for secretKey, secretValue := range secret {
			for _, key := range keys {
				if secretKey == key {
					log.RegisterSecret(secretValue)
					envVariable := VaultCredentialEnvPrefixDefault + ConvertEnvVar(secretKey)
					log.Entry().Debugf("Exposing general purpose credential '%v' as '%v'", key, envVariable)
					os.Setenv(envVariable, secretValue)

					log.RegisterSecret(piperutils.EncodeString(secretValue))
					envVariable = VaultCredentialEnvPrefixDefault + ConvertEnvVar(secretKey) + "_BASE64"
					log.Entry().Debugf("Exposing general purpose base64 encoded credential '%v' as '%v'", key, envVariable)
					os.Setenv(envVariable, piperutils.EncodeString(secretValue))
					matched = true
				}
			}
		}
	}
	return
}

func getTestCredentialKeys(config *StepConfig) []string {
	keysRaw, ok := config.Config[vaultTestCredentialKeys].([]interface{})
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(keysRaw))
	for _, keyRaw := range keysRaw {
		key, ok := keyRaw.(string)
		if !ok {
			log.Entry().Warnf("%s needs to be an array of strings", vaultTestCredentialKeys)
			return nil
		}
		keys = append(keys, key)
	}
	return keys
}

func getCredentialKeys(config *StepConfig) []string {
	keysRaw, ok := config.Config[vaultCredentialKeys].([]interface{})
	if !ok {
		log.Entry().Debugf("Not fetching general purpose credentials from vault since they are not (properly) configured")
		return nil
	}
	keys := make([]string, 0, len(keysRaw))
	for _, keyRaw := range keysRaw {
		key, ok := keyRaw.(string)
		if !ok {
			log.Entry().Warnf("%s is needs to be an array of strings", vaultCredentialKeys)
			return nil
		}
		keys = append(keys, key)
	}
	return keys
}

// ConvertEnvVar converts to a valid environment variable string
func ConvertEnvVar(s string) string {
	r := strings.ToUpper(s)
	r = strings.ReplaceAll(r, "-", "_")
	reg, err := regexp.Compile("[^a-zA-Z0-9_]*")
	if err != nil {
		log.Entry().Debugf("could not compile regex of convertEnvVar: %v", err)
	}
	replacedString := reg.ReplaceAllString(r, "")
	return replacedString
}

// RemoveVaultSecretFiles removes all secret files that have been created during execution
func RemoveVaultSecretFiles() {
	if VaultSecretFileDirectory != "" {
		os.RemoveAll(VaultSecretFileDirectory)
	}
}

func createTemporarySecretFile(namePattern string, content string) (string, error) {
	if VaultSecretFileDirectory == "" {
		var err error
		fileUtils := &piperutils.Files{}
		VaultSecretFileDirectory, err = fileUtils.TempDir("", "vault")
		if err != nil {
			return "", err
		}
	}

	file, err := os.CreateTemp(VaultSecretFileDirectory, namePattern)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

func lookupPath(client VaultClient, path string, param *StepParameters) *string {
	log.Entry().Debug("Checking vault path for secret")
	secret, err := client.GetKvSecret(path)
	if err != nil {
		log.Entry().WithError(err).WithField("vaultPath", path).Warn("Failed to fetch secret from vault")
		return nil
	}
	if secret == nil {
		return nil
	}

	field := secret[param.Name]
	if field != "" {
		log.RegisterSecret(field)
		return &field
	}
	log.Entry().Debugf("Secret did not contain a field name '%s'", param.Name)
	// try parameter aliases
	for _, alias := range param.Aliases {
		log.Entry().Debugf("Trying alias field name '%s'", alias.Name)
		field := secret[alias.Name]
		if field != "" {
			log.RegisterSecret(field)
			if alias.Deprecated {
				log.Entry().WithField("package", "SAP/jenkins-library/pkg/config").Warningf("DEPRECATION NOTICE: old step config key '%s' used in Vault. Please switch to '%s'!", alias.Name, param.Name)
			}
			return &field
		}
	}
	return nil
}

func getSecretReferencePaths(reference *ResourceReference, config map[string]interface{}) []string {
	retPaths := make([]string, 0, len(VaultRootPaths))
	secretName := reference.Default
	if providedName, ok := config[reference.Name].(string); ok && providedName != "" {
		secretName = providedName
	}
	for _, rootPath := range VaultRootPaths {
		fullPath := path.Join(rootPath, secretName)
		retPaths = append(retPaths, fullPath)
	}
	return retPaths
}
