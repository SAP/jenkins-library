package config

import (
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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

		// Central Build in Jenkins was renamed to Build.
		handleLegacyStageNaming(config, currentOrchestrator, stageName)

		// Check #1: Apply explicit activation/deactivation from config file (if any)
		// and then evaluate stepActive conditions
		runStep := make(map[string]bool, len(stage.Steps))
		stepConfigCache := make(map[string]StepConfig, len(stage.Steps))
		for _, step := range stage.Steps {
			// Consider only orchestrator-specific steps if the orchestrator limitation is set.
			if len(step.Orchestrators) > 0 && !slices.Contains(step.Orchestrators, currentOrchestrator) {
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
			return false, fmt.Errorf("only one config key allowed per condition but %v provided", len(s.Config))
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
			return false, fmt.Errorf("failed to check filePattern condition: %w", err)
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
			return false, fmt.Errorf("failed to check filePatternFromConfig condition: %w", err)
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

func checkForNpmScriptsInPackagesV1(npmScript string, config StepConfig, utils piperutils.FileUtils) (bool, error) {
	packages, err := utils.Glob("**/package.json")
	if err != nil {
		return false, fmt.Errorf("failed to check if file-exists: %w", err)
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
			return false, fmt.Errorf("failed to open file %s: %v", pack, err)
		}
		packageJSON := map[string]interface{}{}
		if err := json.Unmarshal(jsonFile, &packageJSON); err != nil {
			return false, fmt.Errorf("failed to unmarshal json file %s: %v", pack, err)
		}
		npmScripts, ok := packageJSON["scripts"]
		if !ok {
			continue
		}
		scriptsMap, ok := npmScripts.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("failed to read scripts from package.json: %T", npmScripts)
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

func handleLegacyStageNaming(c *Config, orchestrator, stageName string) {
	if orchestrator == "Jenkins" && stageName == "Build" {
		_, buildExists := c.Stages["Build"]
		centralBuildStageConfig, centralBuildExists := c.Stages["Central Build"]
		if buildExists && centralBuildExists {
			log.Entry().Warnf("You have 2 entries for build stage in config.yml. " +
				"Parameters defined under 'Central Build' are ignored. " +
				"Please use only 'Build'")
			return
		}

		if centralBuildExists {
			c.Stages["Build"] = centralBuildStageConfig
			log.Entry().Warnf("You are using 'Central Build' stage in config.yml. " +
				"Please move parameters under the 'Build' stage, " +
				"since 'Central Build' will be removed in future releases")
		}
	}
}
