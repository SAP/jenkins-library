package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/piperenv"

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
	Name            string  `json:"name"`
	Aliases         []Alias `json:"aliases,omitempty"`
	Description     string  `json:"description"`
	LongDescription string  `json:"longDescription,omitempty"`
}

// StepSpec defines the spec details for a step, like step inputs, containers, sidecars, ...
type StepSpec struct {
	Inputs     StepInputs  `json:"inputs,omitempty"`
	Outputs    StepOutputs `json:"outputs,omitempty"`
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
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	LongDescription string              `json:"longDescription,omitempty"`
	ResourceRef     []ResourceReference `json:"resourceRef,omitempty"`
	Scope           []string            `json:"scope"`
	Type            string              `json:"type"`
	Mandatory       bool                `json:"mandatory,omitempty"`
	Default         interface{}         `json:"default,omitempty"`
	PossibleValues  []interface{}       `json:"possibleValues,omitempty"`
	Aliases         []Alias             `json:"aliases,omitempty"`
	Conditions      []Condition         `json:"conditions,omitempty"`
	Secret          bool                `json:"secret,omitempty"`
}

// ResourceReference defines the parameters of a resource reference
type ResourceReference struct {
	Name  string `json:"name"`
	Param string `json:"param"`
}

// Alias defines a step input parameter alias
type Alias struct {
	Name       string `json:"name,omitempty"`
	Deprecated bool   `json:"deprecated,omitempty"`
}

// StepResources defines the resources to be provided by the step context, e.g. Jenkins pipeline
type StepResources struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Type        string                   `json:"type,omitempty"`
	Parameters  []map[string]interface{} `json:"params,omitempty"`
	Conditions  []Condition              `json:"conditions,omitempty"`
}

// StepSecrets defines the secrets to be provided by the step context, e.g. Jenkins pipeline
type StepSecrets struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Type        string  `json:"type,omitempty"`
	Aliases     []Alias `json:"aliases,omitempty"`
}

// StepOutputs defines the outputs of a step step, typically one or multiple resources
type StepOutputs struct {
	Resources []StepResources `json:"resources,omitempty"`
}

// Container defines an execution container
type Container struct {
	//ToDo: check dockerOptions, dockerVolumeBind, containerPortMappings, sidecarOptions, sidecarVolumeBind
	Command         []string    `json:"command"`
	EnvVars         []EnvVar    `json:"env"`
	Image           string      `json:"image"`
	ImagePullPolicy string      `json:"imagePullPolicy"`
	Name            string      `json:"name"`
	ReadyCommand    string      `json:"readyCommand"`
	Shell           string      `json:"shell"`
	WorkingDir      string      `json:"workingDir"`
	Conditions      []Condition `json:"conditions,omitempty"`
	Options         []Option    `json:"options,omitempty"`
	//VolumeMounts    []VolumeMount `json:"volumeMounts,omitempty"`
}

// ToDo: Add the missing Volumes part to enable the volume mount completely
// VolumeMount defines a mount path
// type VolumeMount struct {
//	MountPath string `json:"mountPath"`
//	Name      string `json:"name"`
//}

// Option defines an docker option
type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EnvVar defines an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Condition defines an condition which decides when the parameter, resource or container is valid
type Condition struct {
	ConditionRef string  `json:"conditionRef"`
	Params       []Param `json:"params"`
}

// Param defines the parameters serving as inputs to the condition
type Param struct {
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
	filters := StepFilters{All: []string{"verbose"}, General: []string{"verbose"}, Steps: []string{"verbose"}, Stages: []string{"verbose"}, Parameters: []string{"verbose"}}
	for _, param := range m.Spec.Inputs.Parameters {
		parameterKeys := []string{param.Name}
		for _, condition := range param.Conditions {
			for _, dependentParam := range condition.Params {
				parameterKeys = append(parameterKeys, dependentParam.Value)
			}
		}
		filters.All = append(filters.All, parameterKeys...)
		for _, scope := range param.Scope {
			switch scope {
			case "GENERAL":
				filters.General = append(filters.General, parameterKeys...)
			case "STEPS":
				filters.Steps = append(filters.Steps, parameterKeys...)
			case "STAGES":
				filters.Stages = append(filters.Stages, parameterKeys...)
			case "PARAMETERS":
				filters.Parameters = append(filters.Parameters, parameterKeys...)
			case "ENV":
				filters.Env = append(filters.Env, parameterKeys...)
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
		parameterKeys := []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}
		for _, container := range m.Spec.Containers {
			for _, condition := range container.Conditions {
				for _, dependentParam := range condition.Params {
					parameterKeys = append(parameterKeys, dependentParam.Value)
					parameterKeys = append(parameterKeys, dependentParam.Name)
				}
			}
		}
		// ToDo: append dependentParam.Value & dependentParam.Name only according to correct parameter scope and not generally
		containerFilters = append(containerFilters, parameterKeys...)
	}
	if len(m.Spec.Sidecars) > 0 {
		//ToDo: support fallback for "dockerName" configuration property -> via aliasing?
		containerFilters = append(containerFilters, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}...)
		//ToDo: add condition param.Value and param.Name to filter as for Containers
	}
	if len(containerFilters) > 0 {
		filters.All = append(filters.All, containerFilters...)
		filters.General = append(filters.General, containerFilters...)
		filters.Steps = append(filters.Steps, containerFilters...)
		filters.Stages = append(filters.Stages, containerFilters...)
		filters.Parameters = append(filters.Parameters, containerFilters...)
	}
	return filters
}

// GetContextDefaults retrieves context defaults like container image, name, env vars, resources, ...
// It only supports scenarios with one container and optionally one sidecar
func (m *StepData) GetContextDefaults(stepName string) (io.ReadCloser, error) {

	//ToDo error handling empty Containers/Sidecars
	//ToDo handle empty Command
	root := map[string]interface{}{}
	if len(m.Spec.Containers) > 0 {
		for _, container := range m.Spec.Containers {
			key := ""
			if len(container.Conditions) > 0 {
				key = container.Conditions[0].Params[0].Value
			}
			p := map[string]interface{}{}
			if key != "" {
				root[key] = p
			} else {
				p = root
			}
			if len(container.Command) > 0 {
				p["containerCommand"] = container.Command[0]
			}
			p["containerName"] = container.Name
			p["containerShell"] = container.Shell
			p["dockerEnvVars"] = envVarsAsStringSlice(container.EnvVars)
			p["dockerImage"] = container.Image
			p["dockerName"] = container.Name
			p["dockerPullImage"] = container.ImagePullPolicy != "Never"
			p["dockerWorkspace"] = container.WorkingDir
			p["dockerOptions"] = optionsAsStringSlice(container.Options)
			//p["dockerVolumeBind"] = volumeMountsAsStringSlice(container.VolumeMounts)

			// Ready command not relevant for main runtime container so far
			//p[] = container.ReadyCommand
		}

	}

	if len(m.Spec.Sidecars) > 0 {
		if len(m.Spec.Sidecars[0].Command) > 0 {
			root["sidecarCommand"] = m.Spec.Sidecars[0].Command[0]
		}
		root["sidecarEnvVars"] = envVarsAsStringSlice(m.Spec.Sidecars[0].EnvVars)
		root["sidecarImage"] = m.Spec.Sidecars[0].Image
		root["sidecarName"] = m.Spec.Sidecars[0].Name
		root["sidecarPullImage"] = m.Spec.Sidecars[0].ImagePullPolicy != "Never"
		root["sidecarReadyCommand"] = m.Spec.Sidecars[0].ReadyCommand
		root["sidecarWorkspace"] = m.Spec.Sidecars[0].WorkingDir
		root["sidecarOptions"] = optionsAsStringSlice(m.Spec.Sidecars[0].Options)
		//root["sidecarVolumeBind"] = volumeMountsAsStringSlice(m.Spec.Sidecars[0].VolumeMounts)
	}

	// not filled for now since this is not relevant in Kubernetes case
	//root["containerPortMappings"] = m.Spec.Sidecars[0].

	if len(m.Spec.Inputs.Resources) > 0 {
		keys := []string{}
		resources := map[string][]string{}
		for _, resource := range m.Spec.Inputs.Resources {
			if resource.Type == "stash" {
				key := ""
				if len(resource.Conditions) > 0 {
					key = resource.Conditions[0].Params[0].Value
				}
				if resources[key] == nil {
					keys = append(keys, key)
					resources[key] = []string{}
				}
				resources[key] = append(resources[key], resource.Name)
			}
		}

		for _, key := range keys {
			if key == "" {
				root["stashContent"] = resources[""]
			} else {
				if root[key] == nil {
					root[key] = map[string]interface{}{
						"stashContent": resources[key],
					}
				} else {
					p := root[key].(map[string]interface{})
					p["stashContent"] = resources[key]
				}
			}
		}
	}

	c := Config{
		Steps: map[string]map[string]interface{}{
			stepName: root,
		},
	}

	JSON, err := yaml.Marshal(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create context defaults")
	}

	r := ioutil.NopCloser(bytes.NewReader(JSON))
	return r, nil
}

// GetResourceParameters retrieves parameters from a named pipeline resource with a defined path
func (m *StepData) GetResourceParameters(path, name string) map[string]interface{} {
	resourceParams := map[string]interface{}{}

	for _, param := range m.Spec.Inputs.Parameters {
		for _, res := range param.ResourceRef {
			if res.Name == name {
				if val := piperenv.GetParameter(filepath.Join(path, name), res.Param); len(val) > 0 {
					resourceParams[param.Name] = val
				}
			}
		}
	}

	return resourceParams
}

func envVarsAsStringSlice(envVars []EnvVar) []string {
	e := []string{}
	for _, v := range envVars {
		e = append(e, fmt.Sprintf("%v=%v", v.Name, v.Value))
	}
	return e
}

func optionsAsStringSlice(options []Option) []string {
	e := []string{}
	for _, v := range options {
		e = append(e, fmt.Sprintf("%v %v", v.Name, v.Value))
	}
	return e
}

//ToDo: Enable this when the Volumes part is also implemented
//func volumeMountsAsStringSlice(volumeMounts []VolumeMount) []string {
//	e := []string{}
//	for _, v := range volumeMounts {
//		e = append(e, fmt.Sprintf("%v:%v", v.Name, v.MountPath))
//	}
//	return e
//}
