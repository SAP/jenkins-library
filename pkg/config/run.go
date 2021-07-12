package config

import (
	"io"
	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/helper"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/bmatcuk/doublestar"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

// RunConfig ...
type RunConfig struct {
	ConditionFilePath    string
	Conditions           RunConditions
	DeactivateStage      map[string]bool
	DeactivateStageSteps map[string]map[string]bool
}

// RunConditions ...
type RunConditions struct {
	// StageConditions map[string]StepConditions `json:"stages,omitempty"`
	StageConditions map[string]map[string]PipelineConditions `json:"stages,omitempty"`
}

// PipelineConditions ..
type PipelineConditions struct {
	Conditions []v1beta1.PipelineTaskCondition `json:"conditions,omitempty"`
}

// Retriever ...
type Retriever interface {
	GetStepConfig(
		flagValues map[string]interface{},
		paramJSON string,
		configuration io.ReadCloser,
		defaults []io.ReadCloser,
		ignoreCustomDefaults bool,
		filters StepFilters,
		parameters []StepParameters,
		secrets []StepSecrets,
		envParameters map[string]interface{},
		stageName string,
		stepName string,
		stepAliases []Alias,
	) (StepConfig, error)
}

// InitRunConfig ...
func (r *RunConfig) InitRunConfig(config Retriever, stages map[string]map[string]interface{}, filters map[string]StepFilters, parameters map[string][]StepParameters, secrets map[string][]StepSecrets, stepAliases map[string][]Alias, glob func(pattern string) (matches []string, err error)) error {
	if glob == nil {
		glob = doublestar.Glob
	}
	r.DeactivateStage = map[string]bool{}
	r.DeactivateStageSteps = map[string]map[string]bool{}

	if len(r.Conditions.StageConditions) == 0 {
		if err := r.loadConditions(); err != nil {
			return errors.Wrap(err, "failed to load pipeline run conditions")
		}
	}

	err := r.evaluateConditions(config, filters, parameters, secrets, stepAliases)
	if err != nil {
		log.Entry().Errorf("Failed to evaluate step conditions: %v", err)
	}

	return nil
}

// ToDo: optimize parameter handling
func (r *RunConfig) getStepConfig(config Retriever, stageName, stepName string, filters map[string]StepFilters, parameters map[string][]StepParameters, secrets map[string][]StepSecrets, stepAliases map[string][]Alias) (StepConfig, error) {
	// no support for flag values and envParameters
	// so far not considered necessary

	flagValues := map[string]interface{}{} // args of step from pipeline_generated.yml

	envParameters := map[string]interface{}{}

	// parameters via paramJSON not supported
	// not considered releavant for pipeline yaml syntax resolution
	paramJSON := ""

	return config.GetStepConfig(flagValues, paramJSON, nil, nil, false, filters[stepName], parameters[stepName], secrets[stepName], envParameters, stageName, stepName, stepAliases[stepName])
}

func (r *RunConfig) deactivateStageStep(stageName, stepName string) {
	// todo: refactor logic of deactivate stage: check if stage is empty after condition eval?
	// r.DeactivateStage[stageName] = true
	if r.DeactivateStageSteps == nil {
		r.DeactivateStageSteps = map[string]map[string]bool{}
	}
	if r.DeactivateStageSteps[stageName] == nil {
		r.DeactivateStageSteps[stageName] = map[string]bool{}
	}
	r.DeactivateStageSteps[stageName][stepName] = true
}

func (r *RunConfig) loadConditions() error {
	if r.ConditionFilePath == "" {
		return errors.Errorf("empty conditions file path: could not load stage conditions")
	}
	conditionFile, err := helper.OpenFile(r.ConditionFilePath, helper.GithubCredentials)
	if err != nil {
		return errors.Errorf("cannot open stash settings: %v", err)
	}

	defer conditionFile.Close()
	content, err := ioutil.ReadAll(conditionFile)
	if err != nil {
		return errors.Wrapf(err, "error reading %v", r.ConditionFilePath)
	}

	err = yaml.Unmarshal(content, &r.Conditions)
	if err != nil {
		return errors.Errorf("format of configuration is invalid %q: %v", content, err)
	}
	return nil
}

func stepConfigLookup(m map[string]interface{}, stepName, key string) interface{} {
	// flat map: key is on top level
	if m[key] != nil {
		return m[key]
	}
	// lookup for step config with following format
	// general:
	//   <key>: <value>
	// stages:
	//   <stepName>:
	//     <key>: <value>
	// steps:
	//   <stepName>:
	//     <key>: <value>
	if m["general"] != nil {
		general := m["general"].(map[string]interface{})
		if general[key] != nil {
			return general[key]
		}
	}
	if m["stages"] != nil {
		stages := m["stages"].(map[string]interface{})
		if stages[stepName] != nil {
			stageStepConfig := stages[stepName].(map[string]interface{})
			if stageStepConfig[key] != nil {
				return stageStepConfig[key]
			}
		}
	}
	if m["steps"] != nil {
		steps := m["steps"].(map[string]interface{})
		if steps[stepName] != nil {
			stepConfig := steps[stepName].(map[string]interface{})
			if stepConfig[key] != nil {
				return stepConfig[key]
			}
		}
	}
	return nil
}
