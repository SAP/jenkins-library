package config

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Config defines the structure of the config files
type Config struct {
	CustomDefaults []string                          `json:"customDefaults,omitempty"`
	General        map[string]interface{}            `json:"general"`
	Stages         map[string]map[string]interface{} `json:"stages"`
	Steps          map[string]map[string]interface{} `json:"steps"`
	openFile       func(s string) (io.ReadCloser, error)
}

// StepConfig defines the structure for merged step configuration
type StepConfig struct {
	Config map[string]interface{}
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
		return NewParseError(fmt.Sprintf("error unmarshalling %q: %v", content, err))
	}
	return nil
}

// ApplyAliasConfig adds configuration values available on aliases to primary configuration parameters
func (c *Config) ApplyAliasConfig(parameters []StepParameters, filters StepFilters, stageName, stepName string) {
	for _, p := range parameters {
		c.General = setParamValueFromAlias(c.General, filters.General, p)
		if c.Stages[stageName] != nil {
			c.Stages[stageName] = setParamValueFromAlias(c.Stages[stageName], filters.Stages, p)
		}
		if c.Steps[stepName] != nil {
			c.Steps[stepName] = setParamValueFromAlias(c.Steps[stepName], filters.Steps, p)
		}
	}
}

func setParamValueFromAlias(configMap map[string]interface{}, filter []string, p StepParameters) map[string]interface{} {
	if configMap != nil && configMap[p.Name] == nil && sliceContains(filter, p.Name) {
		for _, a := range p.Aliases {
			aliasVal := getDeepAliasValue(configMap, a.Name)
			if aliasVal != nil {
				configMap[p.Name] = aliasVal
			}
			if configMap[p.Name] != nil {
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
		return getDeepAliasValue(configMap[parts[0]].(map[string]interface{}), strings.Join(parts[1:], "/"))
	}
	return configMap[key]
}

// GetStepConfig provides merged step configuration using defaults, config, if available
func (c *Config) GetStepConfig(flagValues map[string]interface{}, paramJSON string, configuration io.ReadCloser, defaults []io.ReadCloser, filters StepFilters, parameters []StepParameters, envParameters map[string]interface{}, stageName, stepName string) (StepConfig, error) {
	var stepConfig StepConfig
	var d PipelineDefaults

	if configuration != nil {
		if err := c.ReadConfig(configuration); err != nil {
			return StepConfig{}, errors.Wrap(err, "failed to parse custom pipeline configuration")
		}
	}
	c.ApplyAliasConfig(parameters, filters, stageName, stepName)

	// consider custom defaults defined in config.yml
	if c.CustomDefaults != nil && len(c.CustomDefaults) > 0 {
		if c.openFile == nil {
			c.openFile = OpenPiperFile
		}
		for _, f := range c.CustomDefaults {
			fc, err := c.openFile(f)
			if err != nil {
				return StepConfig{}, errors.Wrapf(err, "getting default '%v' failed", f)
			}
			defaults = append(defaults, fc)
		}
	}

	if err := d.ReadPipelineDefaults(defaults); err != nil {
		switch err.(type) {
		case *ParseError:
			return StepConfig{}, errors.Wrap(err, "failed to parse pipeline default configuration")
		default:
			//ignoring unavailability of defaults since considered optional
		}
	}

	// initialize with defaults from step.yaml
	stepConfig.mixInStepDefaults(parameters)

	// read defaults & merge general -> steps (-> general -> steps ...)
	for _, def := range d.Defaults {
		def.ApplyAliasConfig(parameters, filters, stageName, stepName)
		stepConfig.mixIn(def.General, filters.General)
		stepConfig.mixIn(def.Steps[stepName], filters.Steps)
	}

	// merge parameters provided by Piper environment
	stepConfig.mixIn(envParameters, filters.All)

	// read config & merge - general -> steps -> stages
	stepConfig.mixIn(c.General, filters.General)
	stepConfig.mixIn(c.Steps[stepName], filters.Steps)
	stepConfig.mixIn(c.Stages[stageName], filters.Stages)

	// merge parameters provided via env vars
	stepConfig.mixIn(envValues(filters.All), filters.All)

	// if parameters are provided in JSON format merge them
	if len(paramJSON) != 0 {
		var params map[string]interface{}
		json.Unmarshal([]byte(paramJSON), &params)

		//apply aliases
		for _, p := range parameters {
			params = setParamValueFromAlias(params, filters.Parameters, p)
		}

		stepConfig.mixIn(params, filters.Parameters)
	}

	// merge command line flags
	if flagValues != nil {
		stepConfig.mixIn(flagValues, filters.Parameters)
	}

	// finally do the condition evaluation post processing
	for _, p := range parameters {
		if len(p.Conditions) > 0 {
			cp := p.Conditions[0].Params[0]
			dependentValue := stepConfig.Config[cp.Name]
			if cmp.Equal(dependentValue, cp.Value) && stepConfig.Config[p.Name] == nil {
				subMapValue := stepConfig.Config[dependentValue.(string)].(map[string]interface{})[p.Name]
				if subMapValue != nil {
					stepConfig.Config[p.Name] = subMapValue
				} else {
					stepConfig.Config[p.Name] = p.Default
				}
			}
		}
	}
	return stepConfig, nil
}

// GetStepConfigWithJSON provides merged step configuration using a provided stepConfigJSON with additional flags provided
func GetStepConfigWithJSON(flagValues map[string]interface{}, stepConfigJSON string, filters StepFilters) StepConfig {
	var stepConfig StepConfig

	stepConfigMap := map[string]interface{}{}

	json.Unmarshal([]byte(stepConfigJSON), &stepConfigMap)

	stepConfig.mixIn(stepConfigMap, filters.All)

	// ToDo: mix in parametersJSON

	if flagValues != nil {
		stepConfig.mixIn(flagValues, filters.Parameters)
	}
	return stepConfig
}

// GetJSON returns JSON representation of an object
func GetJSON(data interface{}) (string, error) {

	result, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrapf(err, "error marshalling json: %v", err)
	}
	return string(result), nil
}

// OpenPiperFile provides functionality to retrieve configuration via file or http
func OpenPiperFile(name string) (io.ReadCloser, error) {
	//ToDo: support also https as source
	if !strings.HasPrefix(name, "http") {
		return os.Open(name)
	}
	return nil, fmt.Errorf("file location not yet supported for '%v'", name)
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

func (s *StepConfig) mixInStepDefaults(stepParams []StepParameters) {
	if s.Config == nil {
		s.Config = map[string]interface{}{}
	}

	for _, p := range stepParams {
		if p.Default != nil {
			s.Config[p.Name] = p.Default
		}
	}
}

func filterMap(data map[string]interface{}, filter []string) map[string]interface{} {
	result := map[string]interface{}{}

	if data == nil {
		data = map[string]interface{}{}
	}

	for key, value := range data {
		if len(filter) == 0 || sliceContains(filter, key) {
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
		if elem == find {
			return true
		}
	}
	return false
}
