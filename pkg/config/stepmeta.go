package config

import (
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// StepData defines the metadata for a step, like step descriptions, parameters, ...
type StepData struct {
	Metadata StepMetadata `json:"metadata"`
	Spec     StepSpec     `json:"spec"`
}

// StepMetadata defines the metadata for a step, like step descriptions, parameters, ...
type StepMetadata struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	LongDescription string `json:"longDescription,omitempty"`
}

// StepSpec defines the spec details for a step, like step inputs, containers, sidecars, ...
type StepSpec struct {
	Inputs StepInputs `json:"inputs"`
	//	Outputs string `json:"description,omitempty"`
	Containers []StepContainers `json:"containers,omitempty"`
	Sidecars   []StepSidecars   `json:"sidecars,omitempty"`
}

// StepInputs defines the spec details for a step, like step inputs, containers, sidecars, ...
type StepInputs struct {
	Parameters []StepParameters `json:"params"`
	Resources  []StepResources  `json:"resources,omitempty"`
	Secrets    []StepSecrets    `json:"secrets,omitempty"`
}

// StepParameters defines the parameters for a step
type StepParameters struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	LongDescription string      `json:"longDescription,omitempty"`
	Scope           []string    `json:"scope"`
	Type            string      `json:"type"`
	Mandatory       bool        `json:"mandatory,omitempty"`
	Default         interface{} `json:"default,omitempty"`
}

// StepResources defines the resources to be provided by the step context, e.g. Jenkins pipeline
type StepResources struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// StepSecrets defines the secrets to be provided by the step context, e.g. Jenkins pipeline
type StepSecrets struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// StepOutputs defines the outputs of a step
//type StepOutputs struct {
//	Name          string `json:"name"`
//}

// StepContainers defines the containers required for a step
type StepContainers struct {
	Containers map[string]interface{} `json:"containers"`
}

// StepSidecars defines any sidears required for a step
type StepSidecars struct {
	Sidecars map[string]interface{} `json:"sidecars"`
}

// StepFilters defines the filter parameters for the different sections
type StepFilters struct {
	All        []string
	General    []string
	Stages     []string
	Steps      []string
	Parameters []string
	Env        []string
}

// ReadPipelineStepData loads step definition in yaml format
func (m *StepData) ReadPipelineStepData(metadata io.ReadCloser) error {
	defer metadata.Close()
	content, err := ioutil.ReadAll(metadata)
	if err != nil {
		return errors.Wrapf(err, "error reading %v", metadata)
	}

	err = yaml.Unmarshal(content, &m)
	if err != nil {
		return errors.Wrapf(err, "error unmarshalling: %v", err)
	}
	return nil
}

// GetParameterFilters retrieves all scope dependent parameter filters
func (m *StepData) GetParameterFilters() StepFilters {
	var filters StepFilters
	for _, param := range m.Spec.Inputs.Parameters {
		for _, scope := range param.Scope {
			filters.All = append(filters.All, param.Name)
			switch scope {
			case "GENERAL":
				filters.General = append(filters.General, param.Name)
			case "STEPS":
				filters.Steps = append(filters.Steps, param.Name)
			case "STAGES":
				filters.Stages = append(filters.Stages, param.Name)
			case "PARAMETERS":
				filters.Parameters = append(filters.Parameters, param.Name)
			case "ENV":
				filters.Env = append(filters.Env, param.Name)
			}
		}
	}
	return filters
}

// GetContextParameterFilters retrieves all scope dependent parameter filters
func (m *StepData) GetContextParameterFilters() StepFilters {
	var filters StepFilters
	for _, secret := range m.Spec.Inputs.Secrets {
		filters.All = append(filters.All, secret.Name)
		filters.General = append(filters.General, secret.Name)
		filters.Steps = append(filters.Steps, secret.Name)
		filters.Stages = append(filters.Stages, secret.Name)
		filters.Parameters = append(filters.Parameters, secret.Name)
		filters.Env = append(filters.Env, secret.Name)
	}
	return filters
}
