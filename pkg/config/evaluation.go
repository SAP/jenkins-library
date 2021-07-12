package config

import (
	"github.com/pkg/errors"
)

// EvaluateConditions validates stage conditions and updates runConfig step deactivation
func (r *RunConfig) evaluateConditions(config Retriever, filters map[string]StepFilters, parameters map[string][]StepParameters, secrets map[string][]StepSecrets, stepAliases map[string][]Alias) error {
	for stageName, stageConditions := range r.Conditions.StageConditions {
		for stepName, stepCondition := range stageConditions {
			stepConfig, err := r.getStepConfig(config, stageName, stepName, filters, parameters, secrets, stepAliases)
			for _, pipelineTaskCondition := range stepCondition.Conditions {
				if err != nil {
					return errors.Wrapf(err, "EvaluateConditions: failed to get stepConfig")
				}
				validationTrue, err := validateCondition(pipelineTaskCondition, stepConfig, stepName)
				if err != nil {
					return errors.Wrapf(err, "EvaluateConditions: validation failed")
				}
				if !validationTrue {
					r.deactivateStageStep(stageName, stepName)
				}
			}
		}
	}
	return nil
}
