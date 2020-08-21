package config

import (
	"github.com/SAP/jenkins-library/pkg/config/interpolation"
	"github.com/SAP/jenkins-library/pkg/vault"
	"github.com/hashicorp/vault/api"

	"github.com/SAP/jenkins-library/pkg/log"
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
		log.Entry().Error("vault or token not set")
		return nil, nil
	}

	// namespaces are only available in vault enterprise so using them should be optional
	namespace := config.Config["vaultNamespace"].(string)

	client, err := vault.NewClient(&api.Config{Address: address}, token, namespace)
	if err != nil {
		log.Entry().Errorf("Creating vault client failed")
		return nil, err
	}

	log.Entry().Errorf("connecting to vault at %s", address)
	log.Entry().Errorf("connecting to namespace %s", namespace)

	return &client, nil
}

func addVaultCredentials(config *StepConfig, client vaultClient, params []StepParameters) error {
	log.Entry().Errorf("vault-debug: in resolve credentials: %#v", params)
	for _, param := range params {

		// we don't overwrite secrets that have already been set in any way
		if v, ok := config.Config[param.Name].(string); ok {
			log.Entry().Errorf("skipping %s for vault lookup since it is already set to %s", param.Name, v)
			continue
		}
		ref := param.GetReference("vaultSecret")
		log.Entry().Errorf("ref is %#v", ref)
		if ref == nil {
			continue
		}
		for _, vaultPath := range ref.Paths {
			log.Entry().Errorf("Resolveing param %s", param.Name)
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
