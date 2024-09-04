package config

import (
	"errors"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/trustengine"
)

// const RefTypeTrustengineSecretFile = "trustengineSecretFile"
const RefTypeTrustengineSecret = "trustengineSecret"

// resolveAllTrustEngineReferences retrieves all the step's secrets from the Trust Engine
func resolveAllTrustEngineReferences(config *StepConfig, params []StepParameters, trustEngineConfiguration trustengine.Configuration, client *piperhttp.Client) {
	for _, param := range params {
		if ref := param.GetReference(RefTypeTrustengineSecret); ref != nil {
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
				log.Entry().Debugf("Skipping getting '%s' from Trust Engine: parameter already set", param.Name)
			}
		}
	}
}

// setTrustEngineConfiguration sets the server URL for the Trust Engine by taking it from the hooks
func (c *Config) setTrustEngineConfiguration(hookConfig map[string]interface{}) error {
	trustEngineHook, ok := hookConfig["trustengine"].(map[string]interface{})
	if !ok {
		return errors.New("no Trust Engine hook configuration found")
	}
	if serverURL, ok := trustEngineHook["serverURL"].(string); ok {
		c.trustEngineConfiguration.ServerURL = serverURL
	} else {
		return errors.New("no server URL found in Trust Engine hook configuration")
	}
	if tokenEndPoint, ok := trustEngineHook["tokenEndPoint"].(string); ok {
		c.trustEngineConfiguration.TokenEndPoint = tokenEndPoint
	} else {
		return errors.New("no token end point found in Trust Engine hook configuration")
	}

	if len(c.trustEngineConfiguration.Token) == 0 {
		log.Entry().Debug("Trust Engine token is not configured or empty string")
	}
	return nil
}

// SetTrustEngineToken sets the token for the Trust Engine
func (c *Config) SetTrustEngineToken(token string) {
	c.trustEngineConfiguration.Token = token
}
