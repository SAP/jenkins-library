package config

import (
	"errors"
	"os"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/trustengine"
)

// ResolveAllTrustEngineReferences retrieves all the step's secrets from the trust engine
func ResolveAllTrustEngineReferences(config *StepConfig, params []StepParameters, trustEngineConfiguration trustengine.Configuration, client *piperhttp.Client) {
	for _, param := range params {
		if ref := param.GetReference("trustengineSecret"); ref != nil {
			if config.Config[param.Name] == "" { // what if Jenkins set the secret?
				log.Entry().Infof("Getting %s from Trust Engine", ref.Name)
				token, err := trustengine.GetToken(ref.Name, client, trustEngineConfiguration)
				if err != nil {
					log.Entry().Info(" failed")
					log.Entry().WithError(err).Warnf("Couldn't get %s secret from Trust Engine", param.Name)
					continue
				}
				log.RegisterSecret(token)
				config.Config[param.Name] = token
				log.Entry().Info(" succeeded")
			}
		}
	}
}

// SetTrustEngineConfiguration sets the server URL and token
func (c *Config) SetTrustEngineConfiguration(hookConfig map[string]interface{}) error {
	c.trustEngineConfiguration = trustengine.Configuration{}
	c.trustEngineConfiguration.Token = os.Getenv("PIPER_TRUST_ENGINE_TOKEN")
	if len(c.trustEngineConfiguration.Token) == 0 {
		log.Entry().Warn("No Trust Engine token environment variable set or is empty string")
	}

	trustEngineHook, ok := hookConfig["trustengine"].(map[string]interface{})
	if !ok {
		return errors.New("no trust engine hook configuration found")
	}
	serverURL, ok := trustEngineHook["serverURL"].(string)
	if ok {
		c.trustEngineConfiguration.ServerURL = serverURL
	} else {
		return errors.New("no server URL found in trust engine hook configuration")
	}
	return nil
}
