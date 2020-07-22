package config

import (
	"path"

	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"
)

// vaultClient interface for mocking
type vaultClient interface {
	GetKvSecret(string) (map[string]string, error)
}

func getVaultClientFromConfig(config StepConfig) (vaultClient, error) {
	address, addressOk := config.Config["vaultAddress"].(string)
	token, tokenOk := config.Config["vaultToken"].(string)

	// if vault isn't used it's not an error
	if !addressOk || !tokenOk {
		return nil, nil
	}

	// namespaces are only available in vault enterprise so using them should be optional
	namespace := config.Config["vaultNamespace"].(string)

	client, err := vault.NewClient(&api.Config{Address: address}, token, namespace)
	if err != nil {
		return nil, err
	}

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
			basePath := ""
			var ok bool
			p, ok := config.Config["vaultBasePath"].(string)
			if ok {
				basePath = p
			}

			secret, err := client.GetKvSecret(path.Join(basePath, vaultPath))
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
