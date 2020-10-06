package config

import (
	"github.com/SAP/jenkins-library/pkg/config/interpolation"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"
)

var vaultFilter = []string{
	"vaultApproleID",
	"vaultApproleSecreId",
	"vaultServerUrl",
	"vaultNamespace",
	"vaultBasePath",
	"vaultPipelineName",
}

// VaultCredentials hold all the auth information needed to fetch configuration from vault
type VaultCredentials struct {
	AppRoleID       string
	AppRoleSecretID string
}

// vaultClient interface for mocking
type vaultClient interface {
	GetKvSecret(string) (map[string]string, error)
}

func getVaultClientFromConfig(config StepConfig, creds VaultCredentials) (vaultClient, error) {
	address, addressOk := config.Config["vaultServerUrl"].(string)
	// if vault isn't used it's not an error
	if !addressOk || creds.AppRoleID == "" || creds.AppRoleSecretID == "" {
		log.Entry().Info("Skipping fetching secrets from vault since it is not configured")
		return nil, nil
	}
	namespace := ""
	// namespaces are only available in vault enterprise so using them should be optional
	if config.Config["vaultNamespace"] != nil {
		namespace = config.Config["vaultNamespace"].(string)
	}

	client, err := vault.NewClientWithAppRole(&api.Config{Address: address}, creds.AppRoleID, creds.AppRoleSecretID, namespace)
	if err != nil {
		return nil, err
	}

	log.Entry().Infof("Fetching secrets from vault at %s", address)
	return &client, nil
}

func addVaultCredentials(config *StepConfig, client vaultClient, params []StepParameters) {
	for _, param := range params {
		// we don't overwrite secrets that have already been set in any way
		if _, ok := config.Config[param.Name].(string); ok {
			continue
		}
		ref := param.GetReference("vaultSecret")
		if ref == nil {
			continue
		}
		for _, vaultPath := range ref.Paths {
			// it should be possible to configure the root path were the secret is stored
			var err error
			vaultPath, err = interpolation.ResolveString(vaultPath, config.Config)
			if err != nil {
				continue
			}

			val := lookupPath(client, vaultPath, &param)
			if val != nil {
				config.Config[param.Name] = *val
			}
		}
	}
}

func lookupPath(client vaultClient, path string, param *StepParameters) *string {
	secret, err := client.GetKvSecret(path)
	if err != nil {
		log.Entry().WithError(err).Warnf("Couldn't fetch secret at %s", path)
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

	// try parameter aliases
	for _, alias := range param.Aliases {
		field := secret[param.Name]
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
