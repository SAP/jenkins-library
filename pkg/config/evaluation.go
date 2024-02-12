package config

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/pkg/errors"
)

const (
	configCondition                = "config"
	configKeysCondition            = "configKeys"
	filePatternFromConfigCondition = "filePatternFromConfig"
	filePatternCondition           = "filePattern"
	npmScriptsCondition            = "npmScripts"
)

// evaluateConditionsV1 validates stage conditions and updates runSteps in runConfig according to V1 schema.
// Priority of step activation/deactivation is follow:
// - stepNotActiveCondition (highest, if any)
// - explicit activation/deactivation (medium, if any)
// - stepActiveConditions (lowest, step is active by default if no conditions are configured)
func (r *RunConfigV1) evaluateConditionsV1(config *Config, utils piperutils.FileUtils, envRootPath string) error {
	if r.RunSteps == nil {
		r.RunSteps = make(map[string]map[string]bool, len(r.PipelineConfig.Spec.Stages))
	}
	if r.RunStages == nil {
		r.RunStages = make(map[string]bool, len(r.PipelineConfig.Spec.Stages))
	}

	currentOrchestrator := orchestrator.DetectOrchestrator().String()
	for _, stage := range r.PipelineConfig.Spec.Stages {
		// Currently, the displayName is being used, but it may be necessary
		// to also consider using the technical name.
		stageName := stage.DisplayName

		// Check #1: Apply explicit activation/deactivation from config file (if any)
		// and then evaluate stepActive conditions
		runStep := make(map[string]bool, len(stage.Steps))
		stepConfigCache := make(map[string]StepConfig, len(stage.Steps))
		for _, step := range stage.Steps {
			// Consider only orchestrator-specific steps if the orchestrator limitation is set.
			if len(step.Orchestrators) > 0 && !piperutils.ContainsString(step.Orchestrators, currentOrchestrator) {
				continue
			}

			stepConfig, err := r.getStepConfig(config, stageName, step.Name, nil, nil, nil, nil)
			if err != nil {
				return err
			}
			stepConfigCache[step.Name] = stepConfig

			// Respect explicit activation/deactivation if available.
			// Note that this has higher priority than step conditions
			if active, ok := stepConfig.Config[step.Name].(bool); ok {
				runStep[step.Name] = active
				continue
			}

			// If no condition is available, the step will be active by default.
			stepActive := true
			for _, condition := range step.Conditions {
				stepActive, err = condition.evaluateV1(stepConfig, utils, step.Name, envRootPath, runStep)
				if err != nil {
					return fmt.Errorf("failed to evaluate step conditions: %w", err)
				}
				if stepActive {
					// The first condition that matches will be considered to activate the step.
					break
				}
			}

			runStep[step.Name] = stepActive
		}

		// Check #2: Evaluate stepNotActive conditions (if any) and deactivate the step if the condition is met.
		//
		// TODO: PART 1 : if explicit activation/de-activation is available should notActiveConditions be checked ?
		// Fortify has no anchor, so if we explicitly set it to true then it may run even during commit pipelines, if we implement TODO PART 1??
		for _, step := range stage.Steps {
			stepConfig, found := stepConfigCache[step.Name]
			if !found {
				// If no stepConfig exists here, it means that this step was skipped in previous checks.
				continue
			}

			for _, condition := range step.NotActiveConditions {
				stepNotActive, err := condition.evaluateV1(stepConfig, utils, step.Name, envRootPath, runStep)
				if err != nil {
					return fmt.Errorf("failed to evaluate not active step conditions: %w", err)
				}

				// Deactivate the step if the notActive condition is met.
				if stepNotActive {
					runStep[step.Name] = false
					break
				}
			}
		}

		r.RunSteps[stageName] = runStep

		stageActive := false
		for _, anyStepIsActive := range r.RunSteps[stageName] {
			if anyStepIsActive {
				stageActive = true
			}
		}
		r.RunStages[stageName] = stageActive
	}

	return nil
}

func (s *StepCondition) evaluateV1(
	config StepConfig,
	utils piperutils.FileUtils,
	stepName string,
	envRootPath string,
	runSteps map[string]bool,
) (bool, error) {

	// only the first condition will be evaluated.
	// if multiple conditions should be checked they need to provided via the Conditions list
	if s.Config != nil {

		if len(s.Config) > 1 {
			return false, errors.Errorf("only one config key allowed per condition but %v provided", len(s.Config))
		}

		// for loop will only cover first entry since we throw an error in case there is more than one config key defined already above
		for param, activationValues := range s.Config {
			for _, activationValue := range activationValues {
				if activationValue == config.Config[param] {
					return true, nil
				}
			}
			return false, nil
		}
	}

	if len(s.ConfigKey) > 0 {
		configKey := strings.Split(s.ConfigKey, "/")
		return checkConfigKeyV1(config.Config, configKey)
	}

	if len(s.FilePattern) > 0 {
		files, err := utils.Glob(s.FilePattern)
		if err != nil {
			return false, errors.Wrap(err, "failed to check filePattern condition")
		}
		if len(files) > 0 {
			return true, nil
		}
		return false, nil
	}

	if len(s.FilePatternFromConfig) > 0 {

		configValue := fmt.Sprint(config.Config[s.FilePatternFromConfig])
		if len(configValue) == 0 {
			return false, nil
		}
		files, err := utils.Glob(configValue)
		if err != nil {
			return false, errors.Wrap(err, "failed to check filePatternFromConfig condition")
		}
		if len(files) > 0 {
			return true, nil
		}
		return false, nil
	}

	if len(s.NpmScript) > 0 {
		return checkForNpmScriptsInPackagesV1(s.NpmScript, config, utils)
	}

	if s.CommonPipelineEnvironment != nil {

		var metadata StepData
		for param, value := range s.CommonPipelineEnvironment {
			cpeEntry := getCPEEntry(param, value, &metadata, stepName, envRootPath)
			if cpeEntry[stepName] == value {
				return true, nil
			}
		}
		return false, nil
	}

	if len(s.PipelineEnvironmentFilled) > 0 {

		var metadata StepData
		param := s.PipelineEnvironmentFilled
		// check CPE for both a string and non-string value
		cpeEntry := getCPEEntry(param, "", &metadata, stepName, envRootPath)
		if len(cpeEntry) == 0 {
			cpeEntry = getCPEEntry(param, nil, &metadata, stepName, envRootPath)
		}

		if _, ok := cpeEntry[stepName]; ok {
			return true, nil
		}

		return false, nil
	}

	if s.OnlyActiveStepInStage {
		// Used only in NotActiveConditions.
		// Returns true if all other steps are inactive, so step will be deactivated
		// if it's the only active step in stage.
		// For example, sapCumulusUpload step must be deactivated in a stage where others steps are inactive.
		return !anyOtherStepIsActive(stepName, runSteps), nil
	}

	// needs to be checked last:
	// if none of the other conditions matches, step will be active unless set to inactive
	if s.Inactive == true {
		return false, nil
	} else {
		return true, nil
	}
}

func getCPEEntry(param string, value interface{}, metadata *StepData, stepName string, envRootPath string) map[string]interface{} {
	dataType := "interface"
	_, ok := value.(string)
	if ok {
		dataType = "string"
	}
	metadata.Spec.Inputs.Parameters = []StepParameters{
		{Name: stepName,
			Type:        dataType,
			ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: param}},
		},
	}
	return metadata.GetResourceParameters(envRootPath, "commonPipelineEnvironment")
}

func checkConfigKeyV1(config map[string]interface{}, configKey []string) (bool, error) {
	value, ok := config[configKey[0]]
	if len(configKey) == 1 {
		return ok, nil
	}
	castedValue, ok := value.(map[string]interface{})
	if !ok {
		return false, nil
	}
	return checkConfigKeyV1(castedValue, configKey[1:])
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

func checkForNpmScriptsInPackagesV1(npmScript string, config StepConfig, utils piperutils.FileUtils) (bool, error) {
	packages, err := utils.Glob("**/package.json")
	if err != nil {
		return false, errors.Wrap(err, "failed to check if file-exists")
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

		jsonFile, err := utils.FileRead(pack)
		if err != nil {
			return false, errors.Errorf("failed to open file %s: %v", pack, err)
		}
		packageJSON := map[string]interface{}{}
		if err := json.Unmarshal(jsonFile, &packageJSON); err != nil {
			return false, errors.Errorf("failed to unmarshal json file %s: %v", pack, err)
		}
		npmScripts, ok := packageJSON["scripts"]
		if !ok {
			continue
		}
		scriptsMap, ok := npmScripts.(map[string]interface{})
		if !ok {
			return false, errors.Errorf("failed to read scripts from package.json: %T", npmScripts)
		}
		if _, ok := scriptsMap[npmScript]; ok {
			return true, nil
		}
	}
	return false, nil
}

// anyOtherStepIsActive loops through previous steps active states and returns true
// if at least one of them is active, otherwise result is false. Ignores the step that is being checked.
func anyOtherStepIsActive(targetStep string, runSteps map[string]bool) bool {
	for step, isActive := range runSteps {
		if isActive && step != targetStep {
			return true
		}
	}

	return false
}
