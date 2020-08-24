package config

import (
	"github.com/SAP/jenkins-library/pkg/config/interpolation"
	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"

	"github.com/SAP/jenkins-library/pkg/log"
)

var vaultFilter = []string{
	"vaultApproleID",
	"vaultApproleSecreId",
	"vaultAddress",
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
	address, addressOk := config.Config["vaultAddress"].(string)
	log.Entry().Infof("config received %#v", config.Config)
	// if vault isn't used it's not an error
	if !addressOk || creds.AppRoleID == "" || creds.AppRoleSecretID == "" {
		log.Entry().Info("Skipping fetching secrets from vault since it is not configured")
		return nil, nil
	}

	// namespaces are only available in vault enterprise so using them should be optional
	namespace := config.Config["vaultNamespace"].(string)

	client, err := vault.NewClientWithAppRole(&api.Config{Address: address}, creds.AppRoleID, creds.AppRoleSecretID, namespace)
	if err != nil {
		log.Entry().Errorf("Creating vault client failed")
		return nil, err
	}

	log.Entry().Infof("Fetching secrets from vault at %s", address)
	return &client, nil
}

func addVaultCredentials(config *StepConfig, client vaultClient, params []StepParameters) error {
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
				return err
			}

			secret, err := client.GetKvSecret(vaultPath)
			if err != nil {
				return err
			}
			if secret == nil {
				continue
			}

			field := secret[param.Name]
			if field != "" {
				config.Config[param.Name] = field
				break
			}
		}
	}
	return nil
}
