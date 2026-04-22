package config

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/piperutils"
	"go.yaml.in/yaml/v3"
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
	Stages map[string]StepConditions `json:"stages,omitempty" yaml:"stages,omitempty"`
}

type StepConditions struct {
	Conditions map[string]map[string]interface{} `json:"stepConditions,omitempty" yaml:"stepConditions,omitempty"`
}

type PipelineDefinitionV1 struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	Spec       Spec     `json:"spec" yaml:"spec"`
	openFile   func(s string, t map[string]string) (io.ReadCloser, error)
	runSteps   map[string]map[string]bool
}

type Metadata struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type Spec struct {
	Stages []Stage `json:"stages" yaml:"stages"`
}

type Stage struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Steps       []Step `json:"steps,omitempty" yaml:"steps,omitempty"`
}

type Step struct {
	Name                string          `json:"name,omitempty" yaml:"name,omitempty"`
	Description         string          `json:"description,omitempty" yaml:"description,omitempty"`
	Conditions          []StepCondition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	NotActiveConditions []StepCondition `json:"notActiveConditions,omitempty" yaml:"notActiveConditions,omitempty"`
	Orchestrators       []string        `json:"orchestrators,omitempty" yaml:"orchestrators,omitempty"`
}

type StepCondition struct {
	Config                    map[string][]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	ConfigKey                 string                   `json:"configKey,omitempty" yaml:"configKey,omitempty"`
	FilePattern               string                   `json:"filePattern,omitempty" yaml:"filePattern,omitempty"`
	FilePatternFromConfig     string                   `json:"filePatternFromConfig,omitempty" yaml:"filePatternFromConfig,omitempty"`
	Inactive                  bool                     `json:"inactive,omitempty" yaml:"inactive,omitempty"`
	OnlyActiveStepInStage     bool                     `json:"onlyActiveStepInStage,omitempty" yaml:"onlyActiveStepInStage,omitempty"`
	NpmScript                 string                   `json:"npmScript,omitempty" yaml:"npmScript,omitempty"`
	CommonPipelineEnvironment map[string]interface{}   `json:"commonPipelineEnvironment,omitempty" yaml:"commonPipelineEnvironment,omitempty"`
	PipelineEnvironmentFilled string                   `json:"pipelineEnvironmentFilled,omitempty" yaml:"pipelineEnvironmentFilled,omitempty"`
}

func (r *RunConfigV1) InitRunConfigV1(config *Config, utils piperutils.FileUtils, envRootPath string) error {

	if len(r.PipelineConfig.Spec.Stages) == 0 {
		if err := r.LoadConditionsV1(); err != nil {
			return fmt.Errorf("failed to load pipeline run conditions: %w", err)
		}
	}

	err := r.evaluateConditionsV1(config, utils, envRootPath)
	if err != nil {
		return fmt.Errorf("failed to evaluate step conditions: %w", err)
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

// LoadConditionsV1 loads stage conditions (in CRD-style) into PipelineConfig
func (r *RunConfigV1) LoadConditionsV1() error {
	defer r.StageConfigFile.Close()
	content, err := io.ReadAll(r.StageConfigFile)
	if err != nil {
		return fmt.Errorf("error: failed to read the stageConfig file: %w", err)
	}

	err = yaml.Unmarshal(content, &r.PipelineConfig)
	if err != nil {
		return fmt.Errorf("format of configuration is invalid %q: %v", content, err)
	}
	return nil
}
