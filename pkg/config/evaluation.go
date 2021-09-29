package config

import (
	"encoding/json"
	"io"
	"path"
	"strings"

	"github.com/pkg/errors"
)

const (
	configCondition                = "config"
	configKeysCondition            = "configKeys"
	filePatternFromConfigCondition = "filePatternFromConfig"
	filePatternCondition           = "filePattern"
	npmScriptsCondition            = "npmScripts"
)

// EvaluateConditionsV2 validates stage conditions and updates runSteps in runConfig according to V1 schema
func (r *RunConfigV1) evaluateConditionsV1(config *Config, filters map[string]StepFilters, parameters map[string][]StepParameters,
	secrets map[string][]StepSecrets, stepAliases map[string][]Alias, glob func(pattern string) (matches []string, err error)) error {
	for _, stage := range r.PipelineConfig.Spec.Stages {
		runStep := map[string]bool{}
		for _, step := range stage.Steps {
			stepActive := false
			stepConfig, err := r.getStepConfig(config, stageName, stepName, filters, parameters, secrets, stepAliases)
			if err != nil {
				return err
			}

			if active, ok := stepConfig.Config[step.Name].(bool); ok {
				// respect explicit activation/de-activation if available
				stepActive = active
			} else {
				for _, condition := range step.Conditions {
					var err error
					switch condition {
					case configCondition:
						if stepActive, err = checkConfig(condition, stepConfig, stepName); err != nil {
							return errors.Wrapf(err, "error: check config condition failed")
						}
					case configKeysCondition:
						if stepActive, err = checkConfigKeys(condition, stepConfig, stepName); err != nil {
							return errors.Wrapf(err, "error: check configKeys condition failed")
						}
					case filePatternFromConfigCondition:
						if stepActive, err = checkForFilesWithPatternFromConfig(condition, stepConfig, stepName, glob); err != nil {
							return errors.Wrapf(err, "error: check filePatternFromConfig condition failed")
						}
					case filePatternCondition:
						if stepActive, err = checkForFilesWithPattern(condition, stepConfig, stepName, glob); err != nil {
							return errors.Wrapf(err, "error: check filePattern condition failed")
						}
					case npmScriptsCondition:
						if stepActive, err = checkForNpmScriptsInPackages(condition, stepConfig, stepName, glob, r.OpenFile); err != nil {
							return errors.Wrapf(err, "error: check npmScripts condition failed")
						}
					default:
						return errors.Errorf("unknown condition %s", conditionName)
					}
					if stepActive {
						break
					}
				}
			}
			runStep[stepName] = stepActive
			r.RunSteps[stageName] = runStep
		}
	}
	return nil
}

// EvaluateConditions validates stage conditions and updates runSteps in runConfig
func (r *RunConfig) evaluateConditions(config *Config, filters map[string]StepFilters, parameters map[string][]StepParameters,
	secrets map[string][]StepSecrets, stepAliases map[string][]Alias, glob func(pattern string) (matches []string, err error)) error {
	for stageName, stepConditions := range r.StageConfig.Stages {
		runStep := map[string]bool{}
		for stepName, stepCondition := range stepConditions.Conditions {
			stepActive := false
			stepConfig, err := r.getStepConfig(config, stageName, stepName, filters, parameters, secrets, stepAliases)
			if err != nil {
				return err
			}

			if active, ok := stepConfig.Config[stepName].(bool); ok {
				// respect explicit activation/de-activation if available
				stepActive = active
			} else {
				for conditionName, condition := range stepCondition {
					var err error
					switch conditionName {
					case configCondition:
						if stepActive, err = checkConfig(condition, stepConfig, stepName); err != nil {
							return errors.Wrapf(err, "error: check config condition failed")
						}
					case configKeysCondition:
						if stepActive, err = checkConfigKeys(condition, stepConfig, stepName); err != nil {
							return errors.Wrapf(err, "error: check configKeys condition failed")
						}
					case filePatternFromConfigCondition:
						if stepActive, err = checkForFilesWithPatternFromConfig(condition, stepConfig, stepName, glob); err != nil {
							return errors.Wrapf(err, "error: check filePatternFromConfig condition failed")
						}
					case filePatternCondition:
						if stepActive, err = checkForFilesWithPattern(condition, stepConfig, stepName, glob); err != nil {
							return errors.Wrapf(err, "error: check filePattern condition failed")
						}
					case npmScriptsCondition:
						if stepActive, err = checkForNpmScriptsInPackages(condition, stepConfig, stepName, glob, r.OpenFile); err != nil {
							return errors.Wrapf(err, "error: check npmScripts condition failed")
						}
					default:
						return errors.Errorf("unknown condition %s", conditionName)
					}
					if stepActive {
						break
					}
				}
			}
			runStep[stepName] = stepActive
			r.RunSteps[stageName] = runStep
		}
	}
	return nil
}

func checkConfig(condition interface{}, config StepConfig, stepName string) (bool, error) {
	switch condition := condition.(type) {
	case string:
		if configValue := stepConfigLookup(config.Config, stepName, condition); configValue != nil {
			return true, nil
		}
	case map[string]interface{}:
		for conditionConfigKey, conditionConfigValue := range condition {
			configValue := stepConfigLookup(config.Config, stepName, conditionConfigKey)
			if configValue == nil {
				return false, nil
			}
			configValueStr, ok := configValue.(string)
			if !ok {
				return false, errors.Errorf("error: config value of %v to compare with is not a string", configValue)
			}
			condConfigValueArr, ok := conditionConfigValue.([]interface{})
			if !ok {
				return false, errors.Errorf("error: type assertion to []interface{} failed: %T", conditionConfigValue)
			}
			for _, item := range condConfigValueArr {
				itemStr, ok := item.(string)
				if !ok {
					return false, errors.Errorf("error: type assertion to string failed: %T", conditionConfigValue)
				}
				if configValueStr == itemStr {
					return true, nil
				}
			}
		}
	default:
		return false, errors.Errorf("error: condidiion type invalid: %T, possible types: string, map[string]interface{}", condition)
	}

	return false, nil
}

func checkConfigKey(configKey string, config StepConfig, stepName string) (bool, error) {
	if configValue := stepConfigLookup(config.Config, stepName, configKey); configValue != nil {
		return true, nil
	}
	return false, nil
}

func checkConfigKeys(condition interface{}, config StepConfig, stepName string) (bool, error) {
	arrCondition, ok := condition.([]interface{})
	if !ok {
		return false, errors.Errorf("error: type assertion to []interface{} failed: %T", condition)
	}
	for _, configKey := range arrCondition {
		if configValue := stepConfigLookup(config.Config, stepName, configKey.(string)); configValue != nil {
			return true, nil
		}
	}
	return false, nil
}

func checkForFilesWithPatternFromConfig(condition interface{}, config StepConfig, stepName string,
	glob func(pattern string) (matches []string, err error)) (bool, error) {
	filePatternConfig, ok := condition.(string)
	if !ok {
		return false, errors.Errorf("error: type assertion to string failed: %T", condition)
	}
	filePatternFromConfig := stepConfigLookup(config.Config, stepName, filePatternConfig)
	if filePatternFromConfig == nil {
		return false, nil
	}
	filePattern, ok := filePatternFromConfig.(string)
	if !ok {
		return false, errors.Errorf("error: type assertion to string failed: %T", filePatternFromConfig)
	}
	matches, err := glob(filePattern)
	if err != nil {
		return false, errors.Wrap(err, "error: failed to check if file-exists")
	}
	if len(matches) > 0 {
		return true, nil
	}
	return false, nil
}

func checkForFilesWithPattern(condition interface{}, config StepConfig, stepName string,
	glob func(pattern string) (matches []string, err error)) (bool, error) {
	switch condition := condition.(type) {
	case string:
		filePattern := condition
		matches, err := glob(filePattern)
		if err != nil {
			return false, errors.Wrap(err, "error: failed to check if file-exists")
		}
		if len(matches) > 0 {
			return true, nil
		}
	case []interface{}:
		filePatterns := condition
		for _, filePattern := range filePatterns {
			filePatternStr, ok := filePattern.(string)
			if !ok {
				return false, errors.Errorf("error: type assertion to string failed: %T", filePatternStr)
			}
			matches, err := glob(filePatternStr)
			if err != nil {
				return false, errors.Wrap(err, "error: failed to check if file-exists")
			}
			if len(matches) > 0 {
				return true, nil
			}
		}
	default:
		return false, errors.Errorf("error: condidiion type invalid: %T, possible types: string, []interface{}", condition)
	}
	return false, nil
}

func checkForNpmScriptsInPackages(condition interface{}, config StepConfig, stepName string,
	glob func(pattern string) (matches []string, err error), openFile func(s string, t map[string]string) (io.ReadCloser, error)) (bool, error) {
	packages, err := glob("**/package.json")
	if err != nil {
		return false, errors.Wrap(err, "error: failed to check if file-exists")
	}
	for _, pack := range packages {
		packDirs := strings.Split(path.Dir(pack), "/")
		isNodeModules := false
		for _, dir := range packDirs {
			if dir == "node_modules" {
				isNodeModules = true
				break
			}
		}
		if isNodeModules {
			continue
		}

		jsonFile, err := openFile(pack, nil)
		if err != nil {
			return false, errors.Errorf("error: failed to open file %s: %v", pack, err)
		}
		defer jsonFile.Close()
		packageJSON := map[string]interface{}{}
		if err := json.NewDecoder(jsonFile).Decode(&packageJSON); err != nil {
			return false, errors.Errorf("error: failed to unmarshal json file %s: %v", pack, err)
		}
		npmScripts, ok := packageJSON["scripts"]
		if !ok {
			continue
		}
		scriptsMap, ok := npmScripts.(map[string]interface{})
		if !ok {
			return false, errors.Errorf("error: type assertion to map[string]interface{} failed: %T", npmScripts)
		}
		switch condition := condition.(type) {
		case string:
			if _, ok := scriptsMap[condition]; ok {
				return true, nil
			}
		case []interface{}:
			for _, conditionNpmScript := range condition {
				conditionNpmScriptStr, ok := conditionNpmScript.(string)
				if !ok {
					return false, errors.Errorf("error: type assertion to string failed: %T", conditionNpmScript)
				}
				if _, ok := scriptsMap[conditionNpmScriptStr]; ok {
					return true, nil
				}
			}
		default:
			return false, errors.Errorf("error: condidiion type invalid: %T, possible types: string, []interface{}", condition)
		}
	}
	return false, nil
}
