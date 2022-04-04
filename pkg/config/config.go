package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

// Config defines the structure of the config files
type Config struct {
	CustomDefaults   []string                          `json:"customDefaults,omitempty"`
	General          map[string]interface{}            `json:"general"`
	Stages           map[string]map[string]interface{} `json:"stages"`
	Steps            map[string]map[string]interface{} `json:"steps"`
	Hooks            map[string]interface{}            `json:"hooks,omitempty"`
	defaults         PipelineDefaults
	initialized      bool
	accessTokens     map[string]string
	openFile         func(s string, t map[string]string) (io.ReadCloser, error)
	vaultCredentials VaultCredentials
}

// StepConfig defines the structure for merged step configuration
type StepConfig struct {
	Config     map[string]interface{}
	HookConfig map[string]interface{}
}

// ReadConfig loads config and returns its content
func (c *Config) ReadConfig(configuration io.ReadCloser) error {
	defer configuration.Close()

	content, err := ioutil.ReadAll(configuration)
	if err != nil {
		return errors.Wrapf(err, "error reading %v", configuration)
	}

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		return NewParseError(fmt.Sprintf("format of configuration is invalid %q: %v", content, err))
	}
	return nil
}

// ApplyAliasConfig adds configuration values available on aliases to primary configuration parameters
func (c *Config) ApplyAliasConfig(parameters []StepParameters, secrets []StepSecrets, filters StepFilters, stageName, stepName string, stepAliases []Alias) {
	// copy configuration from step alias to correct step
	if len(stepAliases) > 0 {
		c.copyStepAliasConfig(stepName, stepAliases)
	}
	for _, p := range parameters {
		c.General = setParamValueFromAlias(stepName, c.General, filters.General, p.Name, p.Aliases)
		if c.Stages[stageName] != nil {
			c.Stages[stageName] = setParamValueFromAlias(stepName, c.Stages[stageName], filters.Stages, p.Name, p.Aliases)
		}
		if c.Steps[stepName] != nil {
			c.Steps[stepName] = setParamValueFromAlias(stepName, c.Steps[stepName], filters.Steps, p.Name, p.Aliases)
		}
	}
	for _, s := range secrets {
		c.General = setParamValueFromAlias(stepName, c.General, filters.General, s.Name, s.Aliases)
		if c.Stages[stageName] != nil {
			c.Stages[stageName] = setParamValueFromAlias(stepName, c.Stages[stageName], filters.Stages, s.Name, s.Aliases)
		}
		if c.Steps[stepName] != nil {
			c.Steps[stepName] = setParamValueFromAlias(stepName, c.Steps[stepName], filters.Steps, s.Name, s.Aliases)
		}
	}
}

func setParamValueFromAlias(stepName string, configMap map[string]interface{}, filter []string, name string, aliases []Alias) map[string]interface{} {
	if configMap != nil && configMap[name] == nil && sliceContains(filter, name) {
		for _, a := range aliases {
			aliasVal := getDeepAliasValue(configMap, a.Name)
			if aliasVal != nil {
				configMap[name] = aliasVal
				if a.Deprecated {
					log.Entry().Warningf("[WARNING] The parameter '%v' is DEPRECATED, use '%v' instead. (%v/%v)", a.Name, name, log.LibraryName, stepName)
				}
			}
			if configMap[name] != nil {
				return configMap
			}
		}
	}
	return configMap
}

func getDeepAliasValue(configMap map[string]interface{}, key string) interface{} {
	parts := strings.Split(key, "/")
	if len(parts) > 1 {
		if configMap[parts[0]] == nil {
			return nil
		}

		paramValueType := reflect.ValueOf(configMap[parts[0]])
		if paramValueType.Kind() != reflect.Map {
			log.Entry().Debugf("Ignoring alias '%v' as '%v' is not pointing to a map.", key, parts[0])
			return nil
		}
		return getDeepAliasValue(configMap[parts[0]].(map[string]interface{}), strings.Join(parts[1:], "/"))
	}
	return configMap[key]
}

func (c *Config) copyStepAliasConfig(stepName string, stepAliases []Alias) {
	for _, stepAlias := range stepAliases {
		if c.Steps[stepAlias.Name] != nil {
			if stepAlias.Deprecated {
				log.Entry().WithField("package", "SAP/jenkins-library/pkg/config").Warningf("DEPRECATION NOTICE: step configuration available for deprecated step '%v'. Please remove or move configuration to step '%v'!", stepAlias.Name, stepName)
			}
			for paramName, paramValue := range c.Steps[stepAlias.Name] {
				if c.Steps[stepName] == nil {
					c.Steps[stepName] = map[string]interface{}{}
				}
				if c.Steps[stepName][paramName] == nil {
					c.Steps[stepName][paramName] = paramValue
				}
			}
		}
	}
}

// InitializeConfig prepares the config object, i.e. loading content, etc.
func (c *Config) InitializeConfig(configuration io.ReadCloser, defaults []io.ReadCloser, ignoreCustomDefaults bool) error {
	if configuration != nil {
		if err := c.ReadConfig(configuration); err != nil {
			return errors.Wrap(err, "failed to parse custom pipeline configuration")
		}
	}

	// consider custom defaults defined in config.yml unless told otherwise
	if ignoreCustomDefaults {
		log.Entry().Info("Ignoring custom defaults from pipeline config")
	} else if c.CustomDefaults != nil && len(c.CustomDefaults) > 0 {
		if c.openFile == nil {
			c.openFile = OpenPiperFile
		}
		for _, f := range c.CustomDefaults {
			fc, err := c.openFile(f, c.accessTokens)
			if err != nil {
				return errors.Wrapf(err, "getting default '%v' failed", f)
			}
			defaults = append(defaults, fc)
		}
	}

	if err := c.defaults.ReadPipelineDefaults(defaults); err != nil {
		return errors.Wrap(err, "failed to read default configuration")
	}
	c.initialized = true
	return nil
}

// GetStepConfig provides merged step configuration using defaults, config, if available
func (c *Config) GetStepConfig(flagValues map[string]interface{}, paramJSON string, configuration io.ReadCloser, defaults []io.ReadCloser, ignoreCustomDefaults bool, filters StepFilters, metadata StepData, envParameters map[string]interface{}, stageName, stepName string) (StepConfig, error) {
	parameters := metadata.Spec.Inputs.Parameters
	secrets := metadata.Spec.Inputs.Secrets
	stepAliases := metadata.Metadata.Aliases

	var stepConfig StepConfig
	var err error

	if !c.initialized {
		err = c.InitializeConfig(configuration, defaults, ignoreCustomDefaults)
		if err != nil {
			return StepConfig{}, err
		}
	}

	c.ApplyAliasConfig(parameters, secrets, filters, stageName, stepName, stepAliases)

	// initialize with defaults from step.yaml
	stepConfig.mixInStepDefaults(parameters)

	// merge parameters provided by Piper environment
	stepConfig.mixIn(envParameters, filters.All)
	stepConfig.mixIn(envParameters, ReportingParameters.getReportingFilter())

	// read defaults & merge general -> steps (-> general -> steps ...)
	for _, def := range c.defaults.Defaults {
		def.ApplyAliasConfig(parameters, secrets, filters, stageName, stepName, stepAliases)
		stepConfig.mixIn(def.General, filters.General)
		stepConfig.mixIn(def.Steps[stepName], filters.Steps)
		stepConfig.mixIn(def.Stages[stageName], filters.Steps)
		stepConfig.mixinVaultConfig(parameters, def.General, def.Steps[stepName], def.Stages[stageName])
		reportingConfig, err := cloneConfig(&def)
		if err != nil {
			return StepConfig{}, err
		}
		reportingConfig.ApplyAliasConfig(ReportingParameters.Parameters, []StepSecrets{}, ReportingParameters.getStepFilters(), stageName, stepName, []Alias{})
		stepConfig.mixinReportingConfig(reportingConfig.General, reportingConfig.Steps[stepName], reportingConfig.Stages[stageName])

		stepConfig.mixInHookConfig(def.Hooks)
	}
	// read hook info from user config file (.pipeline/config.yaml), overwrites hook config from defaults
	stepConfig.mixInHookConfig(c.Hooks)

	// read config & merge - general -> steps -> stages
	stepConfig.mixIn(c.General, filters.General)
	stepConfig.mixIn(c.Steps[stepName], filters.Steps)
	stepConfig.mixIn(c.Stages[stageName], filters.Stages)

	// merge parameters provided via env vars
	stepConfig.mixIn(envValues(filters.All), filters.All)

	// if parameters are provided in JSON format merge them
	if len(paramJSON) != 0 {
		var params map[string]interface{}
		err := json.Unmarshal([]byte(paramJSON), &params)
		if err != nil {
			log.Entry().Warnf("failed to parse parameters from environment: %v", err)
		} else {
			// apply aliases
			for _, p := range parameters {
				params = setParamValueFromAlias(stepName, params, filters.Parameters, p.Name, p.Aliases)
			}
			for _, s := range secrets {
				params = setParamValueFromAlias(stepName, params, filters.Parameters, s.Name, s.Aliases)
			}

			stepConfig.mixIn(params, filters.Parameters)
		}
	}

	// merge command line flags
	if flagValues != nil {
		stepConfig.mixIn(flagValues, filters.Parameters)
	}

	if verbose, ok := stepConfig.Config["verbose"].(bool); ok && verbose {
		log.SetVerbose(verbose)
	} else if !ok && stepConfig.Config["verbose"] != nil {
		log.Entry().Warnf("invalid value for parameter verbose: '%v'", stepConfig.Config["verbose"])
	}

	stepConfig.mixinVaultConfig(parameters, c.General, c.Steps[stepName], c.Stages[stageName])

	reportingConfig, err := cloneConfig(c)
	if err != nil {
		return StepConfig{}, err
	}
	reportingConfig.ApplyAliasConfig(ReportingParameters.Parameters, []StepSecrets{}, ReportingParameters.getStepFilters(), stageName, stepName, []Alias{})
	stepConfig.mixinReportingConfig(reportingConfig.General, reportingConfig.Steps[stepName], reportingConfig.Stages[stageName])

	// check whether vault should be skipped
	if skip, ok := stepConfig.Config["skipVault"].(bool); !ok || !skip {
		// fetch secrets from vault
		vaultClient, err := getVaultClientFromConfig(stepConfig, c.vaultCredentials)
		if err != nil {
			return StepConfig{}, err
		}
		if vaultClient != nil {
			defer vaultClient.MustRevokeToken()
			resolveAllVaultReferences(&stepConfig, vaultClient, append(parameters, ReportingParameters.Parameters...))
			resolveVaultTestCredentials(&stepConfig, vaultClient)
			resolveVaultCredentials(&stepConfig, vaultClient)
		}
	}

	// finally do the condition evaluation post processing
	for _, p := range parameters {
		if len(p.Conditions) > 0 {
			for _, cond := range p.Conditions {
				for _, param := range cond.Params {
					// retrieve configuration value of condition parameter
					dependentValue := stepConfig.Config[param.Name]
					// check if configuration of condition parameter matches the value
					// so far string-equals condition is assumed here
					// if so and if no config applied yet, then try to apply the value
					if cmp.Equal(dependentValue, param.Value) && stepConfig.Config[p.Name] == nil {
						subMap, ok := stepConfig.Config[dependentValue.(string)].(map[string]interface{})
						if ok && subMap[p.Name] != nil {
							stepConfig.Config[p.Name] = subMap[p.Name]
						}
					}
				}
			}
		}
	}
	return stepConfig, nil
}

// SetVaultCredentials sets the appRoleID and the appRoleSecretID or the vaultTokento load additional
// configuration from vault
// Either appRoleID and appRoleSecretID or vaultToken must be specified.
func (c *Config) SetVaultCredentials(appRoleID, appRoleSecretID string, vaultToken string) {
	c.vaultCredentials = VaultCredentials{
		AppRoleID:       appRoleID,
		AppRoleSecretID: appRoleSecretID,
		VaultToken:      vaultToken,
	}
}

// GetStepConfigWithJSON provides merged step configuration using a provided stepConfigJSON with additional flags provided
func GetStepConfigWithJSON(flagValues map[string]interface{}, stepConfigJSON string, filters StepFilters) StepConfig {
	var stepConfig StepConfig

	stepConfigMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(stepConfigJSON), &stepConfigMap)
	if err != nil {
		log.Entry().Warnf("invalid stepConfig JSON: %v", err)
	}

	stepConfig.mixIn(stepConfigMap, filters.All)

	// ToDo: mix in parametersJSON

	if flagValues != nil {
		stepConfig.mixIn(flagValues, filters.Parameters)
	}
	return stepConfig
}

func (c *Config) GetStageConfig(paramJSON string, configuration io.ReadCloser, defaults []io.ReadCloser, ignoreCustomDefaults bool, acceptedParams []string, stageName string) (StepConfig, error) {

	filters := StepFilters{
		General:    acceptedParams,
		Steps:      []string{},
		Stages:     acceptedParams,
		Parameters: acceptedParams,
		Env:        []string{},
	}
	return c.GetStepConfig(map[string]interface{}{}, paramJSON, configuration, defaults, ignoreCustomDefaults, filters, StepData{}, map[string]interface{}{}, stageName, "")
}

// GetJSON returns JSON representation of an object
func GetJSON(data interface{}) (string, error) {

	result, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrapf(err, "error marshalling json: %v", err)
	}
	return string(result), nil
}

// GetYAML returns YAML representation of an object
func GetYAML(data interface{}) (string, error) {

	result, err := yaml.Marshal(data)
	if err != nil {
		return "", errors.Wrapf(err, "error marshalling yaml: %v", err)
	}
	return string(result), nil
}

// OpenPiperFile provides functionality to retrieve configuration via file or http
func OpenPiperFile(name string, accessTokens map[string]string) (io.ReadCloser, error) {
	if len(name) == 0 {
		return nil, errors.Wrap(os.ErrNotExist, "no filename provided")
	}

	if !strings.HasPrefix(name, "http://") && !strings.HasPrefix(name, "https://") {
		return os.Open(name)
	}

	return httpReadFile(name, accessTokens)
}

func httpReadFile(name string, accessTokens map[string]string) (io.ReadCloser, error) {

	u, err := url.Parse(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read url: %w", err)
	}

	// support http(s) urls next to file path
	client := piperhttp.Client{}

	var header http.Header
	if len(accessTokens[u.Host]) > 0 {
		client.SetOptions(piperhttp.ClientOptions{Token: fmt.Sprintf("token %v", accessTokens[u.Host])})
		header = map[string][]string{"Accept": {"application/vnd.github.v3.raw"}}
	}

	response, err := client.SendRequest("GET", name, nil, header, nil)
	if err != nil {
		return nil, err
	}
	return response.Body, nil
}

func envValues(filter []string) map[string]interface{} {
	vals := map[string]interface{}{}
	for _, param := range filter {
		if envVal := os.Getenv("PIPER_" + param); len(envVal) != 0 {
			vals[param] = os.Getenv("PIPER_" + param)
		}
	}
	return vals
}

func (s *StepConfig) mixIn(mergeData map[string]interface{}, filter []string) {

	if s.Config == nil {
		s.Config = map[string]interface{}{}
	}

	s.Config = merge(s.Config, filterMap(mergeData, filter))
}

func (s *StepConfig) mixInHookConfig(mergeData map[string]interface{}) {

	if s.HookConfig == nil {
		s.HookConfig = map[string]interface{}{}
	}

	s.HookConfig = merge(s.HookConfig, mergeData)
}

func (s *StepConfig) mixInStepDefaults(stepParams []StepParameters) {
	if s.Config == nil {
		s.Config = map[string]interface{}{}
	}

	// conditional defaults need to be written to a sub map
	// in order to prevent a "last one wins" situation
	// this is then considered at the end of GetStepConfig once the complete configuration is known
	for _, p := range stepParams {
		if p.Default != nil {
			if len(p.Conditions) == 0 {
				s.Config[p.Name] = p.Default
			} else {
				for _, cond := range p.Conditions {
					for _, param := range cond.Params {
						s.Config[param.Value] = map[string]interface{}{p.Name: p.Default}
					}
				}
			}
		}
	}
}

// ApplyContainerConditions evaluates conditions in step yaml container definitions
func ApplyContainerConditions(containers []Container, stepConfig *StepConfig) {
	for _, container := range containers {
		if len(container.Conditions) > 0 {
			for _, param := range container.Conditions[0].Params {
				if container.Conditions[0].ConditionRef == "strings-equal" && stepConfig.Config[param.Name] == param.Value {
					var containerConf map[string]interface{}
					if stepConfig.Config[param.Value] != nil {
						containerConf = stepConfig.Config[param.Value].(map[string]interface{})
						for key, value := range containerConf {
							if stepConfig.Config[key] == nil {
								stepConfig.Config[key] = value
							}
						}
						delete(stepConfig.Config, param.Value)
					}
				}
			}
		}
	}
}

func filterMap(data map[string]interface{}, filter []string) map[string]interface{} {
	result := map[string]interface{}{}

	if data == nil {
		data = map[string]interface{}{}
	}

	for key, value := range data {
		if value != nil && (len(filter) == 0 || sliceContains(filter, key)) {
			result[key] = value
		}
	}
	return result
}

func merge(base, overlay map[string]interface{}) map[string]interface{} {

	result := map[string]interface{}{}

	if base == nil {
		base = map[string]interface{}{}
	}

	for key, value := range base {
		result[key] = value
	}

	for key, value := range overlay {
		if val, ok := value.(map[string]interface{}); ok {
			if valBaseKey, ok := base[key].(map[string]interface{}); !ok {
				result[key] = merge(map[string]interface{}{}, val)
			} else {
				result[key] = merge(valBaseKey, val)
			}
		} else {
			result[key] = value
		}
	}
	return result
}

func sliceContains(slice []string, find string) bool {
	for _, elem := range slice {
		matches, _ := regexp.MatchString(elem, find)
		if matches {
			return true
		}
	}
	return false
}

func cloneConfig(config *Config) (*Config, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	clone := &Config{}
	if err = json.Unmarshal(configJSON, &clone); err != nil {
		return nil, err
	}

	return clone, nil
}
