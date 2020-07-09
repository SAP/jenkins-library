package config

import (
	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"
)

func getVaultClientFromConfig(config StepConfig) (*vault.Client, error) {
	address, addressOk := config.Config["vaultAddress"].(string)
	rootPath, rootPathOk := config.Config["vaultRootPath"].(string)
	token, tokenOk := config.Config["vaultToken"].(string)

	// if vault isn't used it's not an error
	if !addressOk || address == "" || !rootPathOk || rootPath == "" || !tokenOk || token == "" {
		return nil, nil
	}

	// namespaces are only available in vault enterprise so using them should be optional
	namespace := config.Config["vaultNamespace"].(string)

	client, err := vault.NewClient(&api.Config{Address: address}, token, namespace)
	if err != nil {
		return nil, err
	}

	client.BindRootPath(rootPath)
	return &client, nil
}

func getVaultConfig(client *vault.Client, config StepConfig, params []StepParameters) (map[string]interface{}, error) {

	vaultConfig := map[string]interface{}{}
	for _, param := range params {

		// we don't overwrite secrets that have already been set in any way
		if val, ok := config.Config[param.Name]; val == "" || !ok {
			continue
		}
		for _, ref := range param.GetReferences("vaultSecret") {
			// it should be possible to configure the path were the secret is stored
			secretPath := config.Config[ref.Name].(string)
			secret, err := client.GetKvSecret(secretPath)
			if err != nil {
				return nil, err
			}
			if secret == nil {
				continue
			}

			field := secret[param.Name]
			if field != "" {
				vaultConfig[param.Name] = field
				break
			}
		}
	}
	return vaultConfig, nil
}
