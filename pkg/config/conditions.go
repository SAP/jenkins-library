package config

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

const (
	conditionRefConfigEquals = "config-equals"
	conditionRefConfigExists = "config-exists"
	conditionRefFileExists   = "file-exists"
	conditionRefIsActive     = "is-active"

	paramNameConfigKey             = "configKey"
	paramNameStepName              = "stepName"
	paramNameConfigCompare         = "contains"
	paramNameFilePattern           = "filePattern"
	paramNameFilePatternFromConfig = "filePatternFromConfig"
	paramNameIsActive              = "activation"
)

func validateCondition(condition v1beta1.PipelineTaskCondition, stepConfig StepConfig, stepName string) (vaidation bool, err error) {
	switch condition.ConditionRef {
	case conditionRefConfigEquals:
		if len(condition.Params) != 2 {
			return false, errors.Errorf("step condition validation for %v has unsupported number of string equal parameters (2 required, got %v)", stepName, len(condition.Params))
		}
		if condition.Params[0].Value.StringVal == "" ||
			(condition.Params[1].Value.ArrayVal == nil && condition.Params[1].Value.StringVal == "") {
			return false, errors.Errorf("step condition validation config-equals for %v or has unsupported param definition or type", stepName)
		}
		if condition.Params[0].Name != paramNameConfigKey || condition.Params[1].Name != paramNameConfigCompare {
			return false, errors.Errorf("step condition validation config-equals for %v has unsupported param names", stepName)
		}
		return validateConfigEquals(condition.Params[0].Value.StringVal, stepName, condition.Params[1].Value, stepConfig)
	case conditionRefConfigExists:
		if len(condition.Params) != 1 {
			return false, errors.Errorf("step condition validation for %v has unsupported number of string equal parameters (1 required, got %v)", stepName, len(condition.Params))
		}
		if condition.Params[0].Value.StringVal == "" {
			return false, errors.Errorf("step condition validation config-exists for %v has empty value", stepName)
		}
		if condition.Params[0].Name == paramNameConfigKey {
			return validateConfigExists(condition.Params[0].Value.StringVal, stepName, stepConfig)
		} else if condition.Params[0].Name == paramNameStepName {
			return validateConfigExistsStepName(condition.Params[0].Value.StringVal, stepName, stepConfig)
		} else {
			return false, errors.Errorf("step condition validation config-exists for %v has unsupported param name: \"%v\"", stepName, condition.Params[0].Name)
		}
	case conditionRefFileExists:
		if len(condition.Params) != 1 {
			return false, errors.Errorf("step condition validation for %v has unsupported number of string equal parameters (1 required, got %v)", stepName, len(condition.Params))
		}
		if !((condition.Params[0].Name == paramNameFilePattern) || (condition.Params[0].Name == paramNameFilePatternFromConfig)) {
			return false, errors.Errorf("step condition validation file-exists for %v has unsupported param name: \"%v\"", stepName, condition.Params[0].Name)
		}
		if condition.Params[0].Value.StringVal == "" {
			return false, errors.Errorf("step condition validation file-exists for %v has empty value", stepName)
		}
		return validateConfigFilePattern(condition.Params[0], stepConfig, doublestar.Glob)
	case conditionRefIsActive:
		if len(condition.Params) != 1 {
			return false, errors.Errorf("step condition validation for %v has unsupported number of parameters (1 required, got %v)", stepName, len(condition.Params))
		}
		if !(condition.Params[0].Name == paramNameIsActive) {
			return false, errors.Errorf("step condition validation is-active for %v has unsupported param name: \"%v\"", stepName, condition.Params[0].Name)
		}
		if condition.Params[0].Value.StringVal == "" {
			return false, errors.Errorf("step condition validation is-active for %v has empty value", stepName)
		}
		return validateIsActive(condition.Params[0].Value.StringVal, stepName)
	default:
		return false, errors.Errorf("validateCondition error: unknown conditionRef found for %v", stepName)
	}

}

func validateConfigEquals(configKey string, stepName string, compareValues v1beta1.ArrayOrString, config StepConfig) (bool, error) {
	configValue := stepConfigLookup(config.Config, stepName, configKey)
	if configValue == nil {
		// key not in map
		return false, nil
	}
	if reflect.TypeOf(configValue).String() != reflect.String.String() {
		// only string values allowed to be compared
		return false, errors.Errorf("validateConfigEquals error: config value of %v to compare with is not a string", configKey)
	}
	switch compareValues.Type {
	case v1beta1.ParamTypeString:
		if configValue == compareValues.StringVal {
			return true, nil
		}
	case v1beta1.ParamTypeArray:
		for _, compcompareValue := range compareValues.ArrayVal {
			if configValue == compcompareValue {
				return true, nil
			}
		}
	default:
		return false, errors.Errorf("condition validation: unexcpected condition param type")
	}
	return false, nil
}

func validateConfigExists(configKey string, stepName string, config StepConfig) (bool, error) {
	if configValue := stepConfigLookup(config.Config, stepName, configKey); configValue != nil {
		return true, nil
	}
	return false, nil
}

func validateConfigExistsStepName(configKey string, stepName string, config StepConfig) (bool, error) {
	if config.Config["steps"] != nil {
		steps := config.Config["steps"].(map[string]interface{})
		if steps[stepName] != nil {
			return true, nil
		}
	}
	return false, nil
}

func validateConfigFilePattern(filePatternParam v1beta1.Param, stepConfig StepConfig, glob func(pattern string) (matches []string, err error)) (bool, error) {
	filePattern := ""
	if filePatternParam.Name == paramNameFilePattern {
		filePattern = filePatternParam.Value.StringVal
	} else if filePatternParam.Name == paramNameFilePatternFromConfig {
		filePatternFromConfig := stepConfig.Config[filePatternParam.Value.StringVal]
		if filePatternFromConfig == nil {
			return false, nil
		}
		if reflect.TypeOf(filePatternFromConfig).String() != reflect.String.String() {
			return false, errors.Errorf("validateConfigFilePattern error: retrieved value from config is not a string value")
		}
		filePatternFromConfigStr := fmt.Sprintf("%v", filePatternFromConfig)
		filePattern = filePatternFromConfigStr
	}
	matches, err := glob(filePattern)
	if err != nil {
		return false, errors.Wrap(err, "validateConfigFilePattern error: failed to check if file-exists")
	}
	if len(matches) > 0 {
		return true, nil
	}
	return false, nil
}

func validateIsActive(isActiveValue, stepName string) (bool, error) {
	validation, err := strconv.ParseBool(isActiveValue)
	if err != nil {
		return false, errors.Wrapf(err, "validateIsActive failed for step %v", stepName)
	}
	return validation, nil
}
