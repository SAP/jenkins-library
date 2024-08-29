package config

import (
	"fmt"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/vault"
)

func ResolveAllTrustEngineReferences(config *StepConfig, params []StepParameters, trustEngineConfiguration vault.TrustEngineConfiguration) {
	for _, param := range params {
		if ref := param.GetReference("trustEngine"); ref != nil {
			if config.Config[param.Name] == "" {
				resolveTrustEngineReference(ref, config, &piperhttp.Client{}, param, trustEngineConfiguration)
			}
		}
	}
}

// resolveTrustEngineReference retrieves a secret from Vault trust engine
func resolveTrustEngineReference(ref *ResourceReference, config *StepConfig, client *piperhttp.Client, param StepParameters, trustEngineConfiguration vault.TrustEngineConfiguration) {
	token, err := vault.GetTrustEngineSecret(ref.Name, client, trustEngineConfiguration)
	if err != nil {
		log.Entry().Infof(fmt.Sprintf("couldn't get secret from trust engine: %s", err))
		return
	}
	log.RegisterSecret(token)
	config.Config[param.Name] = token
	log.Entry().Infof("retrieving %s token from trust engine succeeded", ref.Name)
}
