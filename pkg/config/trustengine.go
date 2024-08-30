package config

import (
	"fmt"
	"os"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/trustengine"
)

func ResolveAllTrustEngineReferences(config *StepConfig, params []StepParameters, trustEngineConfiguration trustengine.Configuration) {
	for _, param := range params {
		if ref := param.GetReference("trustengineSecret"); ref != nil {
			if config.Config[param.Name] == "" { // what if Jenkins set the secret?
				resolveTrustEngineReference(ref, config, &piperhttp.Client{}, param, trustEngineConfiguration)
			}
		}
	}
}

// resolveTrustEngineReference retrieves a secret from Vault trust engine
func resolveTrustEngineReference(ref *ResourceReference, config *StepConfig, client *piperhttp.Client, param StepParameters, trustEngineConfiguration trustengine.Configuration) {
	token, err := trustengine.GetToken(ref.Name, client, trustEngineConfiguration)
	if err != nil {
		log.Entry().Infof(fmt.Sprintf("couldn't get secret from trust engine: %s", err))
		return
	}
	log.RegisterSecret(token)
	config.Config[param.Name] = token
	log.Entry().Infof("retrieving %s token from trust engine succeeded", ref.Name)
}

// SetTrustEngineConfiguration sets the server URL and token
func (c *Config) SetTrustEngineConfiguration(hookConfig map[string]interface{}) {
	c.trustEngineConfiguration = trustengine.Configuration{}
	c.trustEngineConfiguration.Token = os.Getenv("PIPER_TRUST_ENGINE_TOKEN")

	trustEngineHook, ok := hookConfig["trustEngine"].(map[string]interface{})
	if !ok {
		log.Entry().Debug("no trust engine hook configuration found")
	}
	serverURL, ok := trustEngineHook["serverURL"].(string)
	if ok {
		c.trustEngineConfiguration.ServerURL = serverURL
	} else {
		log.Entry().Debug("no server URL found in trust engine hook configuration")
	}
}
