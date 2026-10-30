package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config/interpolation"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/systemtrust"
)

const RefTypeSystemTrustSecret = "systemTrustSecret"
const skipSystemTrust = "skipSystemTrust"

// resolveAllSystemTrustReferences retrieves all the step's secrets from the System Trust
func resolveAllSystemTrustReferences(config *StepConfig, params []StepParameters, systemTrustConfiguration systemtrust.Configuration, client *piperhttp.Client) {

	skip, _ := config.Config[skipSystemTrust].(bool)
	for _, param := range params {
		if ref := param.GetReference(RefTypeSystemTrustSecret); ref != nil {
			if skip {
				log.Entry().Debugf("Skipping retrieval of '%s' from System Trust: disabled by configuration", param.Name)
				continue
			}
			paramValue, _ := config.Config[param.Name].(string)
			if paramValue == "" {
				log.Entry().Infof("Getting '%s' from System Trust", param.Name)
				refName, args, err := resolveSystemTrustReference(*ref, config.Config)
				if err != nil {
					log.Entry().WithError(err).Warnf("System Trust: failed to resolve reference for '%s'", param.Name)
					continue
				}
				token, err := systemtrust.GetToken(refName, client, systemTrustConfiguration, args)
				if err != nil {
					log.Entry().WithError(err).Warnf("System Trust: failed to retrieve '%s' (key: '%s')", param.Name, ref.Default)
					continue
				}
				if token == "" {
					log.Entry().Warnf("System Trust: failed to retrieve '%s' (key: '%s'): empty token returned", param.Name, ref.Default)
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

var systemTrustParameterPattern = regexp.MustCompile(`\$\(([a-zA-Z0-9\.]*)\)`)

func resolveSystemTrustReference(ref ResourceReference, values map[string]any) (string, map[string]string, error) {
	resolve := func(input, context string) (string, error) {
		for _, match := range systemTrustParameterPattern.FindAllStringSubmatch(input, -1) {
			name := match[1]
			_, ok := values[name].(string)
			if !ok {
				return "", fmt.Errorf("%s: parameter '%s' is missing or not a string", context, name)
			}
		}
		resolved, ok := interpolation.ResolveString(input, values)
		if !ok {
			return "", fmt.Errorf("%s could not be interpolated", context)
		}
		return resolved, nil
	}
	refName, err := resolve(ref.Default, "System Trust reference")
	if err != nil {
		return "", nil, err
	}
	if systemTrustParameterPattern.MatchString(ref.Default) && strings.Contains(refName, "<scope>") {
		parts := strings.SplitN(refName, "<scope>", 2)
		if parts[0] == "" || parts[1] == "" {
			return "", nil, errors.New("System Trust reference requires nonempty system and scope")
		}
	}
	var args map[string]string
	if len(ref.Args) > 0 {
		args = make(map[string]string, len(ref.Args))
		for key, input := range ref.Args {
			value, err := resolve(input, fmt.Sprintf("System Trust argument '%s'", key))
			if err != nil {
				return "", nil, err
			}
			if value == "" {
				return "", nil, fmt.Errorf("System Trust argument '%s' is empty", key)
			}
			args[key] = value
		}
	}
	return refName, args, nil
}

// setSystemTrustConfiguration sets the server URL for the System Trust by taking it from the hooks
func (c *Config) setSystemTrustConfiguration(hookConfig map[string]any) error {
	systemTrustHook, ok := hookConfig["systemtrust"].(map[string]any)
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

	if len(c.systemTrustConfiguration.Token) == 0 {
		return errors.New("no System Trust token found and envvar is empty")
	}
	return nil
}

// SetSystemTrustToken sets the token for the System Trust
func (c *Config) SetSystemTrustToken(token string) {
	c.systemTrustConfiguration.Token = token
}
