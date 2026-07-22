package config

import (
	"errors"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/systemtrust"
)

const RefTypeSystemTrustSecret = "systemTrustSecret"

// resolveAllSystemTrustReferences retrieves all the step's secrets from the System Trust
func resolveAllSystemTrustReferences(config *StepConfig, params []StepParameters, systemTrustConfiguration systemtrust.Configuration, client *piperhttp.Client) {
	for _, param := range params {
		if ref := param.GetReference(RefTypeSystemTrustSecret); ref != nil {
			if config.Config[param.Name] == "" {
				log.Entry().Infof("Getting '%s' from System Trust", param.Name)
				token, err := systemtrust.GetToken(ref.Default, client, systemTrustConfiguration)
				if err != nil {
					log.Entry().WithError(err).Warnf("System Trust: failed to retrieve '%s' (key: '%s')", param.Name, ref.Default)
					continue
				}
				log.RegisterSecret(token)
				config.Config[param.Name] = token
				log.Entry().Info(" succeeded")
			} else {
				log.Entry().Debugf("Skipping retrieval of '%s' from System Trust: parameter already set", param.Name)
			}
		}
	}
}

// setSystemTrustConfiguration sets the server URL for the System Trust by taking it from the hooks.
// If stepConfig contains a "systemTrustURL" value (e.g. set via CPE or pipeline config), it takes
// precedence over the default hook serverURL, allowing users to point to a staging instance.
func (c *Config) setSystemTrustConfiguration(hookConfig map[string]interface{}, stepConfig map[string]interface{}) error {
	systemTrustHook, ok := hookConfig["systemtrust"].(map[string]interface{})
	if !ok {
		return errors.New("no System Trust hook configuration found")
	}
	if serverURL, ok := systemTrustHook["serverURL"].(string); ok {
		c.systemTrustConfiguration.ServerURL = serverURL
	} else {
		return errors.New("no System Trust server URL found")
	}
	// Allow the user to override the default hook serverURL via a "systemTrustURL" parameter
	// (e.g. set in .pipeline/config.yml, via CPE, or as a step parameter).
	if overrideURL, ok := stepConfig["systemTrustURL"].(string); ok && overrideURL != "" {
		log.Entry().Infof("Overriding System Trust server URL with user-provided value: %s", overrideURL)
		c.systemTrustConfiguration.ServerURL = overrideURL
	}
	if tokenEndPoint, ok := systemTrustHook["tokenEndPoint"].(string); ok {
		c.systemTrustConfiguration.TokenEndPoint = tokenEndPoint
	} else {
		return errors.New("no System Trust service endpoint found")
	}
	if tokenQueryParamName, ok := systemTrustHook["tokenQueryParamName"].(string); ok {
		c.systemTrustConfiguration.TokenQueryParamName = tokenQueryParamName
	} else {
		return errors.New("no System Trust query parameter name found")
	}

	if len(c.systemTrustConfiguration.Token) == 0 {
		return errors.New("no System Trust token found and envvar is empty")
	}
	return nil
}

// SetSystemTrustToken sets the token for the System Trust
func (c *Config) SetSystemTrustToken(token string) {
	c.systemTrustConfiguration.Token = token
}
