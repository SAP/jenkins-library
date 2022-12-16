package config

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// RunConfig ...
type RunConfig struct {
	StageConfigFile io.ReadCloser
	StageConfig     StageConfig
	RunStages       map[string]bool
	RunSteps        map[string]map[string]bool
	OpenFile        func(s string, t map[string]string) (io.ReadCloser, error)
	FileUtils       *piperutils.Files
}

type RunConfigV1 struct {
	RunConfig
	PipelineConfig PipelineDefinitionV1
}

type StageConfig struct {
	Stages map[string]StepConditions `json:"stages,omitempty"`
}

type StepConditions struct {
	Conditions map[string]map[string]interface{} `json:"stepConditions,omitempty"`
}

type PipelineDefinitionV1 struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
	openFile   func(s string, t map[string]string) (io.ReadCloser, error)
	runSteps   map[string]map[string]bool
}

type Metadata struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
}

type Spec struct {
	Stages []Stage `json:"stages"`
}

type Stage struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Steps       []Step `json:"steps,omitempty"`
}

type Step struct {
	Name                string          `json:"name,omitempty"`
	Description         string          `json:"description,omitempty"`
	Conditions          []StepCondition `json:"conditions,omitempty"`
	NotActiveConditions []StepCondition `json:"notActiveConditions,omitempty"`
	Orchestrators       []string        `json:"orchestrators,omitempty"`
}

type StepCondition struct {
	Config                    map[string][]interface{} `json:"config,omitempty"`
	ConfigKey                 string                   `json:"configKey,omitempty"`
	FilePattern               string                   `json:"filePattern,omitempty"`
	FilePatternFromConfig     string                   `json:"filePatternFromConfig,omitempty"`
	Inactive                  bool                     `json:"inactive,omitempty"`
	NpmScript                 string                   `json:"npmScript,omitempty"`
	CommonPipelineEnvironment map[string]interface{}   `json:"commonPipelineEnvironment,omitempty"`
	PipelineEnvironmentFilled string                   `json:"pipelineEnvironmentFilled,omitempty"`
}

func (r *RunConfigV1) InitRunConfigV1(config *Config, filters map[string]StepFilters, parameters map[string][]StepParameters,
	secrets map[string][]StepSecrets, stepAliases map[string][]Alias, utils piperutils.FileUtils, envRootPath string) error {

	if len(r.PipelineConfig.Spec.Stages) == 0 {
		if err := r.LoadConditionsV1(); err != nil {
			return fmt.Errorf("failed to load pipeline run conditions: %w", err)
		}
	}

	err := r.evaluateConditionsV1(config, filters, parameters, secrets, stepAliases, utils, envRootPath)
	if err != nil {
		return fmt.Errorf("failed to evaluate step conditions: %w", err)
	}

	return nil
}

// InitRunConfig ...
func (r *RunConfig) InitRunConfig(config *Config, filters map[string]StepFilters, parameters map[string][]StepParameters,
	secrets map[string][]StepSecrets, stepAliases map[string][]Alias, glob func(pattern string) (matches []string, err error),
	openFile func(s string, t map[string]string) (io.ReadCloser, error)) error {
	r.OpenFile = openFile
	r.RunSteps = map[string]map[string]bool{}

	if len(r.StageConfig.Stages) == 0 {
		if err := r.loadConditions(); err != nil {
			return errors.Wrap(err, "failed to load pipeline run conditions")
		}
	}

	err := r.evaluateConditions(config, filters, parameters, secrets, stepAliases, glob)
	if err != nil {
		return errors.Wrap(err, "failed to evaluate step conditions: %v")
	}

	return nil
}

// ToDo: optimize parameter handling
func (r *RunConfig) getStepConfig(config *Config, stageName, stepName string, filters map[string]StepFilters,
	parameters map[string][]StepParameters, secrets map[string][]StepSecrets, stepAliases map[string][]Alias) (StepConfig, error) {
	// no support for flag values and envParameters
	// so far not considered necessary

	flagValues := map[string]interface{}{} // args of step from pipeline_generated.yml

	envParameters := map[string]interface{}{}

	// parameters via paramJSON not supported
	// not considered releavant for pipeline yaml syntax resolution
	paramJSON := ""

	stepMeta := StepData{
		Spec: StepSpec{
			Inputs: StepInputs{Parameters: parameters[stepName], Secrets: secrets[stepName]},
		},
		Metadata: StepMetadata{Aliases: stepAliases[stepName]},
	}

	return config.GetStepConfig(flagValues, paramJSON, nil, nil, false, filters[stepName], stepMeta, envParameters, stageName, stepName)
}

func (r *RunConfig) loadConditions() error {
	defer r.StageConfigFile.Close()
	content, err := ioutil.ReadAll(r.StageConfigFile)
	if err != nil {
		return errors.Wrapf(err, "error: failed to read the stageConfig file")
	}

	err = yaml.Unmarshal(content, &r.StageConfig)
	if err != nil {
		return errors.Errorf("format of configuration is invalid %q: %v", content, err)
	}
	return nil
}

// LoadConditionsV1 loads stage conditions (in CRD-style) into PipelineConfig
func (r *RunConfigV1) LoadConditionsV1() error {
	defer r.StageConfigFile.Close()
	content, err := ioutil.ReadAll(r.StageConfigFile)
	if err != nil {
		return errors.Wrapf(err, "error: failed to read the stageConfig file")
	}

	err = yaml.Unmarshal(content, &r.PipelineConfig)
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
