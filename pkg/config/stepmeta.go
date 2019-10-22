package config

import (
	"bytes"
	"fmt"
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
	Containers []Container `json:"containers,omitempty"`
	Sidecars   []Container `json:"sidecars,omitempty"`
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

// Container defines an execution container
type Container struct {
	//ToDo: check dockerOptions, dockerVolumeBind, containerPortMappings, sidecarOptions, sidecarVolumeBind
	Command         []string `json:"command"`
	EnvVars         []EnvVar `json:"env"`
	Image           string   `json:"image"`
	ImagePullPolicy string   `json:"imagePullPolicy"`
	Name            string   `json:"name"`
	ReadyCommand    string   `json:"readyCommand"`
	Shell           string   `json:"shell"`
	WorkingDir      string   `json:"workingDir"`
}

// EnvVar defines an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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
		filters.All = append(filters.All, param.Name)
		for _, scope := range param.Scope {
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

	containerFilters := []string{}
	if len(m.Spec.Containers) > 0 {
		containerFilters = append(containerFilters, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}...)
	}
	if len(m.Spec.Sidecars) > 0 {
		//ToDo: support fallback for "dockerName" configuration property -> via aliasing?
		containerFilters = append(containerFilters, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}...)
	}
	if len(containerFilters) > 0 {
		filters.All = append(filters.All, containerFilters...)
		filters.Steps = append(filters.Steps, containerFilters...)
		filters.Stages = append(filters.Stages, containerFilters...)
		filters.Parameters = append(filters.Parameters, containerFilters...)
	}
	return filters
}

// GetContextDefaults retrieves context defaults like container image, name, env vars, ...
// It only supports scenarios with one container and optionally one sidecar
func (m *StepData) GetContextDefaults(stepName string) (io.ReadCloser, error) {

	p := map[string]interface{}{}

	//ToDo error handling empty Containers/Sidecars
	//ToDo handle empty Command

	if len(m.Spec.Containers) > 0 {
		if len(m.Spec.Containers[0].Command) > 0 {
			p["containerCommand"] = m.Spec.Containers[0].Command[0]
		}
		p["containerName"] = m.Spec.Containers[0].Name
		p["containerShell"] = m.Spec.Containers[0].Shell
		p["dockerEnvVars"] = envVarsAsStringSlice(m.Spec.Containers[0].EnvVars)
		p["dockerImage"] = m.Spec.Containers[0].Image
		p["dockerName"] = m.Spec.Containers[0].Name
		p["dockerPullImage"] = m.Spec.Containers[0].ImagePullPolicy != "Never"
		p["dockerWorkspace"] = m.Spec.Containers[0].WorkingDir

		// Ready command not relevant for main runtime container so far
		//p[] = m.Spec.Containers[0].ReadyCommand
	}

	if len(m.Spec.Sidecars) > 0 {
		if len(m.Spec.Sidecars[0].Command) > 0 {
			p["sidecarCommand"] = m.Spec.Sidecars[0].Command[0]
		}
		p["sidecarEnvVars"] = envVarsAsStringSlice(m.Spec.Sidecars[0].EnvVars)
		p["sidecarImage"] = m.Spec.Sidecars[0].Image
		p["sidecarName"] = m.Spec.Sidecars[0].Name
		p["sidecarPullImage"] = m.Spec.Sidecars[0].ImagePullPolicy != "Never"
		p["sidecarReadyCommand"] = m.Spec.Sidecars[0].ReadyCommand
		p["sidecarWorkspace"] = m.Spec.Sidecars[0].WorkingDir
	}

	// not filled for now since this is not relevant in Kubernetes case
	//p["dockerOptions"] = m.Spec.Containers[0].
	//p["dockerVolumeBind"] = m.Spec.Containers[0].
	//p["containerPortMappings"] = m.Spec.Sidecars[0].
	//p["sidecarOptions"] = m.Spec.Sidecars[0].
	//p["sidecarVolumeBind"] = m.Spec.Sidecars[0].

	c := Config{
		Steps: map[string]map[string]interface{}{
			stepName: p,
		},
	}

	JSON, err := yaml.Marshal(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create context defaults")
	}

	r := ioutil.NopCloser(bytes.NewReader(JSON))
	return r, nil
}

func envVarsAsStringSlice(envVars []EnvVar) []string {
	e := []string{}
	for _, v := range envVars {
		e = append(e, fmt.Sprintf("%v=%v", v.Name, v.Value))
	}
	return e
}