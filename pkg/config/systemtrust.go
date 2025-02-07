package config

import (
	"errors"
	"fmt"

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
					log.Entry().Info(" failed")
					log.Entry().WithError(err).Debugf("Couldn't get '%s' token from System Trust", ref.Default)
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

// setSystemTrustConfiguration sets the server URL for the System Trust by taking it from the hooks
func (c *Config) setSystemTrustConfiguration(hookConfig map[string]interface{}) error {
	systemTrustHook, ok := hookConfig["systemtrust"].(map[string]interface{})
	if !ok {
		return errors.New("no System Trust hook configuration found")
	}
	if serverURL, ok := systemTrustHook["serverURL"].(string); ok {
		c.systemTrustConfiguration.ServerURL = serverURL
	} else {
		return errors.New("no System Trust server URL found")
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

	if len(c.systemTrustConfiguration.Token) == 0 || c.systemTrustConfiguration.Token == "null" {
		return errors.New("no System Trust token found and envvar is empty")
	}
	return nil
}

// SetSystemTrustToken sets the token for the System Trust
func (c *Config) SetSystemTrustToken(token string) {
	c.systemTrustConfiguration.Token = token
	fmt.Printf("got System trust token: %v\n", token)
	fmt.Printf("Length of token: %d\n", len(token))
	fmt.Printf("Length of token: %T\n", token)
}
