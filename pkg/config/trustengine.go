package config

import (
	"errors"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/trustengine"
)

// resolveAllTrustEngineReferences retrieves all the step's secrets from the trust engine
func resolveAllTrustEngineReferences(config *StepConfig, params []StepParameters, trustEngineConfiguration trustengine.Configuration, client *piperhttp.Client) {
	for _, param := range params {
		if ref := param.GetReference(trustengine.RefTypeSecret); ref != nil {
			if config.Config[param.Name] == "" { // what if Jenkins set the secret?
				log.Entry().Infof("Getting '%s' from Trust Engine", param.Name)
				token, err := trustengine.GetToken(ref.Name, client, trustEngineConfiguration)
				if err != nil {
					log.Entry().Info(" failed")
					log.Entry().WithError(err).Warnf("Couldn't get %s secret from Trust Engine", param.Name)
					continue
				}
				log.RegisterSecret(token)
				config.Config[param.Name] = token
				log.Entry().Info(" succeeded")
			} else {
				log.Entry().Infof("Skipping getting '%s' from Trust Engine: parameter already set", param.Name)
			}
		}
	}
}

// setTrustEngineServer sets the server URL for the Trust Engine
func (c *Config) setTrustEngineServer(hookConfig map[string]interface{}) error {
	trustEngineHook, ok := hookConfig["trustengine"].(map[string]interface{})
	if !ok {
		return errors.New("no trust engine hook configuration found")
	}
	if serverURL, ok := trustEngineHook["serverURL"].(string); ok {
		c.trustEngineConfiguration.ServerURL = serverURL
	} else {
		return errors.New("no server URL found in trust engine hook configuration")
	}
	return nil
}

// SetTrustEngineToken sets the token for the Trust Engine
func (c *Config) SetTrustEngineToken(token string) {
	c.trustEngineConfiguration.Token = token
	if len(c.trustEngineConfiguration.Token) == 0 {
		log.Entry().Warn("Trust Engine token is not configured or empty string")
	}
}
