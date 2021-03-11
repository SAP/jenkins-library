package config

import (
	"io/ioutil"
	"os"

	"github.com/SAP/jenkins-library/pkg/config/interpolation"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"
)

var (
	vaultFilter = []string{
		"vaultAppRoleID",
		"vaultAppRoleSecreId",
		"vaultServerUrl",
		"vaultNamespace",
		"vaultBasePath",
		"vaultPipelineName",
		"vaultPath",
		"skipVault",
		"vaultDisableOverwrite",
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

func (s *StepConfig) mixinVaultConfig(configs ...map[string]interface{}) {
	for _, config := range configs {
		s.mixIn(config, vaultFilter)
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
	for _, vaultPath := range ref.Paths {
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
