package config

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config/interpolation"
	"github.com/SAP/jenkins-library/pkg/log"
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
	vaultTestCredentialEnvPrefixDefault = "PIPER_TESTCREDENTIAL_"
	vaultCredentialEnvPrefixDefault     = "PIPER_VAULTCREDENTIAL_"
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
		vaultCredentialEnvPrefixDefault,
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

// vaultClient interface for mocking
type vaultClient interface {
	GetKvSecret(string) (map[string]string, error)
	MustRevokeToken()
}

func (s *StepConfig) mixinVaultConfig(parameters []StepParameters, configs ...map[string]interface{}) {
	for _, config := range configs {
		s.mixIn(config, vaultFilter)
		// when an empty filter is returned we skip the mixin call since an empty filter will allow everything
		if referencesFilter := getFilterForResourceReferences(parameters); len(referencesFilter) > 0 {
			s.mixIn(config, referencesFilter)
		}
	}
}

func getVaultClientFromConfig(config StepConfig, creds VaultCredentials) (vaultClient, error) {
	address, addressOk := config.Config["vaultServerUrl"].(string)
	// if vault isn't used it's not an error

	if !addressOk || creds.VaultToken == "" && (creds.AppRoleID == "" || creds.AppRoleSecretID == "") {
		log.Entry().Debug("Skipping fetching secrets from vault since it is not configured")
		return nil, nil
	}
	namespace := ""
	// namespaces are only available in vault enterprise so using them should be optional
	if config.Config["vaultNamespace"] != nil {
		namespace = config.Config["vaultNamespace"].(string)
		log.Entry().Debugf("Using vault namespace %s", namespace)
	}

	var client vaultClient
	var err error
	clientConfig := &vault.Config{Config: &api.Config{Address: address}, Namespace: namespace}
	if creds.VaultToken != "" {
		log.Entry().Debugf("Using Vault Token Authentication")
		client, err = vault.NewClient(clientConfig, creds.VaultToken)
	} else {
		log.Entry().Debugf("Using Vaults AppRole Authentication")
		client, err = vault.NewClientWithAppRole(clientConfig, creds.AppRoleID, creds.AppRoleSecretID)
	}
	if err != nil {
		return nil, err
	}

	log.Entry().Infof("Fetching secrets from vault at %s", address)
	return client, nil
}

func resolveAllVaultReferences(config *StepConfig, client vaultClient, params []StepParameters) {
	for _, param := range params {
		if ref := param.GetReference("vaultSecret"); ref != nil {
			resolveVaultReference(ref, config, client, param)
		}
		if ref := param.GetReference("vaultSecretFile"); ref != nil {
			resolveVaultReference(ref, config, client, param)
		}
	}
}

func resolveVaultReference(ref *ResourceReference, config *StepConfig, client vaultClient, param StepParameters) {
	vaultDisableOverwrite, _ := config.Config["vaultDisableOverwrite"].(bool)
	if _, ok := config.Config[param.Name].(string); vaultDisableOverwrite && ok {
		log.Entry().Debugf("Not fetching '%s' from vault since it has already been set", param.Name)
		return
	}

	var secretValue *string
	for _, vaultPath := range getSecretReferencePaths(ref, config.Config) {
		// it should be possible to configure the root path were the secret is stored
		vaultPath, ok := interpolation.ResolveString(vaultPath, config.Config)
		if !ok {
			continue
		}

		secretValue = lookupPath(client, vaultPath, &param)
		if secretValue != nil {
			log.Entry().Debugf("Resolved param '%s' with vault path '%s'", param.Name, vaultPath)
			if ref.Type == "vaultSecret" {
				config.Config[param.Name] = *secretValue
			} else if ref.Type == "vaultSecretFile" {
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
		log.Entry().Warnf("Could not resolve param '%s' from vault", param.Name)
	}
}

// resolve test credential keys and expose as environment variables
func resolveVaultTestCredentials(config *StepConfig, client vaultClient) {
	credPath, pathOk := config.Config[vaultTestCredentialPath].(string)
	keys := getTestCredentialKeys(config)
	if !(pathOk && keys != nil) || credPath == "" || len(keys) == 0 {
		log.Entry().Debugf("Not fetching test credentials from vault since they are not (properly) configured")
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
		secretsResolved = populateTestCredentialsAsEnvs(config, secret, keys)
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
				envVariable := vaultTestCredentialEnvPrefix + convertEnvVar(secretKey)
				log.Entry().Debugf("Exposing test credential '%v' as '%v'", key, envVariable)
				os.Setenv(envVariable, secretValue)
				matched = true
			}
		}
	}
	return
}

// resolve credential keys and expose as environment variables
func resolveVaultCredentials(config *StepConfig, client vaultClient) {
	credPath, pathOk := config.Config[vaultCredentialPath].(string)
	keys := getCredentialKeys(config)
	if !(pathOk && keys != nil) || credPath == "" || len(keys) == 0 {
		log.Entry().Debugf("Not fetching test credentials from vault since they are not (properly) configured")
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

func populateCredentialsAsEnvs(config *StepConfig, secret map[string]string, keys []string) (matched bool) {

	vaultCredentialEnvPrefix, ok := config.Config["vaultCredentialEnvPrefix"].(string)
	isCredentialEnvPrefixDefault := false

	if !ok || len(vaultCredentialEnvPrefix) == 0 {
		vaultCredentialEnvPrefix = vaultCredentialEnvPrefixDefault
		isCredentialEnvPrefixDefault = true
	}
	for secretKey, secretValue := range secret {
		for _, key := range keys {
			if secretKey == key {
				log.RegisterSecret(secretValue)
				envVariable := vaultCredentialEnvPrefix + convertEnvVar(secretKey)
				log.Entry().Debugf("Exposing general purpose credential '%v' as '%v'", key, envVariable)
				os.Setenv(envVariable, secretValue)
				matched = true
			}
		}
	}

	// we always create the env variable with the default prefx so that
	// we can always refer to it in steps if its to be hard-coded
	if !isCredentialEnvPrefixDefault {
		for secretKey, secretValue := range secret {
			for _, key := range keys {
				if secretKey == key {
					log.RegisterSecret(secretValue)
					envVariable := vaultCredentialEnvPrefixDefault + convertEnvVar(secretKey)
					log.Entry().Debugf("Exposing general purpose credential '%v' as '%v'", key, envVariable)
					os.Setenv(envVariable, secretValue)
					matched = true
				}
			}
		}
	}
	return
}

func getCredentialKeys(config *StepConfig) []string {
	keysRaw, ok := config.Config[vaultCredentialKeys].([]interface{})
	if !ok {
		log.Entry().Debugf("Not fetching test credentials from vault since they are not (properly) configured")
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

func getTestCredentialKeys(config *StepConfig) []string {
	keysRaw, ok := config.Config[vaultTestCredentialKeys].([]interface{})
	if !ok {
		log.Entry().Debugf("Not fetching test credentials from vault since they are not (properly) configured")
		return nil
	}
	keys := make([]string, 0, len(keysRaw))
	for _, keyRaw := range keysRaw {
		key, ok := keyRaw.(string)
		if !ok {
			log.Entry().Warnf("%s is needs to be an array of strings", vaultTestCredentialKeys)
			return nil
		}
		keys = append(keys, key)
	}
	return keys
}

// converts to a valid environment variable string
func convertEnvVar(s string) string {
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
		VaultSecretFileDirectory, err = ioutil.TempDir("", "vault")
		if err != nil {
			return "", err
		}
	}

	file, err := ioutil.TempFile(VaultSecretFileDirectory, namePattern)
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

func lookupPath(client vaultClient, path string, param *StepParameters) *string {
	log.Entry().Debugf("Trying to resolve vault parameter '%s' at '%s'", param.Name, path)
	secret, err := client.GetKvSecret(path)
	if err != nil {
		log.Entry().WithError(err).Warnf("Couldn't fetch secret at '%s'", path)
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
				log.Entry().WithField("package", "SAP/jenkins-library/pkg/config").Warningf("DEPRECATION NOTICE: old step config key '%s' used in vault. Please switch to '%s'!", alias.Name, param.Name)
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

func toStringSlice(interfaceSlice []interface{}) []string {
	retSlice := make([]string, 0, len(interfaceSlice))
	for _, vRaw := range interfaceSlice {
		if v, ok := vRaw.(string); ok {
			retSlice = append(retSlice, v)
			continue
		}
		log.Entry().Warnf("'%s' needs to be of type string or an array of strings but got %T (%[2]v)", vaultPath, vRaw)
	}
	return retSlice
}
