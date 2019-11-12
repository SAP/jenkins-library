package helper

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/ghodss/yaml"
)

// ContextDefaultData defines the metadata for a step, like step descriptions, parameters, ...
type ContextDefaultData struct {
	Metadata   ContextDefaultMetadata     `json:"metadata"`
	Parameters []ContextDefaultParameters `json:"params"`
}

// ContextDefaultMetadata defines the metadata for a step, like step descriptions, parameters, ...
type ContextDefaultMetadata struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	LongDescription string `json:"longDescription,omitempty"`
}

// ContextDefaultParameters defines the parameters for a step
type ContextDefaultParameters struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ReadPipelineContextDefaultData loads step definition in yaml format
func (c *ContextDefaultData) readPipelineContextDefaultData(metadata io.ReadCloser) {
	defer metadata.Close()
	content, err := ioutil.ReadAll(metadata)
	checkError(err)
	err = yaml.Unmarshal(content, &c)
	checkError(err)
}

// ReadContextDefaultMap maps the default descriptions into a map
func (c *ContextDefaultData) readContextDefaultMap() map[string]string {
	var m map[string]string = make(map[string]string)

	for _, param := range c.Parameters {
		m[param.Name] = param.Description
	}

	return m
}

func readContextDefaultDescription(contextDefaultPath string) map[string]string {
	//read context default description
	var ContextDefaultData ContextDefaultData

	contextDefaultFile, err := os.Open(contextDefaultPath)
	checkError(err)
	defer contextDefaultFile.Close()

	ContextDefaultData.readPipelineContextDefaultData(contextDefaultFile)
	return ContextDefaultData.readContextDefaultMap()
}

// generates the step documentation and replaces the template with the generated documentation
func generateStepDocumentation(stepData config.StepData, docTemplatePath string) {

	docTemplateFilePath := fmt.Sprintf("%v%v.md", docTemplatePath, stepData.Metadata.Name)

	//check if template exists otherwise print No Template found
	if _, err := os.Stat(docTemplateFilePath); os.IsNotExist(err) {
		fmt.Printf("No Template found for Step: %v \n", stepData.Metadata.Name)
		return
	}

	setDefaultStepParameters(&stepData)

	content := readAndAdjustTemplate(docTemplateFilePath)

	// binding of functions and placeholder
	funcMap := template.FuncMap{
		"docGenDescription":            docGenDescription,
		"docGenStepName":               docGenStepName,
		"docGenParameters":             docGenParameters,
		"docGenConfiguration":          docGenConfiguration,
		"docJenkinsPluginDependencies": docJenkinsPluginDependencies,
	}
	tmpl, err := template.New("doc").Funcs(funcMap).Parse(content)
	checkError(err)

	// add secrets, context defaults to the step parameters
	handleStepParameters(&stepData)

	//overwrite existing file
	docFile, err := os.OpenFile(docTemplateFilePath, os.O_WRONLY, 0644)
	defer docFile.Close()
	checkError(err)

	//write executed template data to the previously opened file
	err = tmpl.Execute(docFile, stepData)
	checkError(err)

	fmt.Printf("Documentation generation complete for: %v\n", stepData.Metadata.Name)
}

func setDefaultStepParameters(stepData *config.StepData) {
	for k, param := range stepData.Spec.Inputs.Parameters {

		if param.Default == nil {
			switch param.Type {
			case "bool":
				param.Default = "false"
			}
		} else {
			switch param.Type {
			case "string":
				param.Default = fmt.Sprintf("\"%v\"", param.Default)
			case "bool":
				boolVal := "false"
				if param.Default.(bool) == true {
					boolVal = "true"
				}
				param.Default = boolVal
			}
		}

		stepData.Spec.Inputs.Parameters[k] = param
	}
}

func readAndAdjustTemplate(docTemplateFilePath string) string {
	fmt.Printf("Open template: %v\n", docTemplateFilePath)
	docFile, err := os.Open(docTemplateFilePath)
	defer docFile.Close()
	checkError(err)

	//read template content
	content, err := ioutil.ReadAll(docFile)
	checkError(err)
	contentStr := string(content)

	//replace old placeholder with new ones
	contentStr = strings.ReplaceAll(contentStr, "${docGenStepName}", "{{docGenStepName .}}")
	contentStr = strings.ReplaceAll(contentStr, "${docGenConfiguration}", "{{docGenConfiguration .}}")
	contentStr = strings.ReplaceAll(contentStr, "${docGenParameters}", "{{docGenParameters .}}")
	contentStr = strings.ReplaceAll(contentStr, "${docGenDescription}", "{{docGenDescription .}}")
	contentStr = strings.ReplaceAll(contentStr, "${docJenkinsPluginDependencies}", "{{docJenkinsPluginDependencies .}}")

	return contentStr
}

//	Replaces the docGenDescription placeholder with content from the yaml
func docGenDescription(stepData config.StepData) string {

	desc := "Description \n\n"

	desc += stepData.Metadata.LongDescription

	return desc
}

// Replaces the docGenStepName placeholder with the content from the yaml
func docGenStepName(stepData config.StepData) string {
	return stepData.Metadata.Name
}

// Replaces the docGenParameters placeholder with the content from the yaml
func docGenParameters(stepData config.StepData) string {
	//create step parameter table
	parametersTable := createParametersTable(stepData.Spec.Inputs.Parameters)
	//create parameters detail section
	parametersDetail := createParametersDetail(stepData.Spec.Inputs.Parameters)

	return "Parameters\n\n" + parametersTable + parametersDetail
}

// Replaces the docGenConfiguration placeholder with the content from the yaml
func docGenConfiguration(stepData config.StepData) string {

	var conf = "We recommend to define values of step parameters via [config.yml file](../configuration.md). \n\n"
	conf += "In following sections of the config.yml the configuration is possible:\n\n"

	// create step configuration table
	conf += createConfigurationTable(stepData.Spec.Inputs.Parameters)

	return conf
}

// Replaces the docGenConfiguration placeholder with default content
func docJenkinsPluginDependencies(stepData config.StepData) string {
	t := "Dependencies \n\n"
	t += "The step depends on the following Jenkins plugins \n\n"
	t += "* &lt;none&gt; \n\n"
	t += "Transitive dependencies are omitted. \n"
	t += " \n"
	t += "The list might be incomplete. \n"
	t += " \n"
	t += "Consider using the [ppiper/jenkins-master](https://cloud.docker.com/u/ppiper/repository/docker/ppiper/jenkins-master) \n"
	t += "docker image. This images comes with preinstalled plugins. \n\n"
	return t
}

func createParametersTable(parameters []config.StepParameters) string {

	var table = "| name | mandatory | default |\n"
	table += "| ---- | --------- | ------- |\n"

	m := combineEqualParametersTogether(parameters)

	for _, param := range parameters {
		if v, ok := m[param.Name]; ok {
			table += fmt.Sprintf(" | %v | %v | %v | \n ", param.Name, ifThenElse(param.Mandatory && param.Default == nil, "Yes", "No"), v)
			delete(m, param.Name)
		}
	}
	return table
}

func createParametersDetail(parameters []config.StepParameters) string {

	var detail = "## Details\n"

	var m map[string]bool = make(map[string]bool)
	for _, param := range parameters {
		if _, ok := m[param.Name]; !ok {
			if len(param.Description) > 0 {
				detail += fmt.Sprintf(" * ` %v ` :  %v \n ", param.Name, param.Description)
				m[param.Name] = true
			}
		}
	}

	return detail
}

//combines equal parameters and the values
func combineEqualParametersTogether(parameters []config.StepParameters) map[string]string {
	var m map[string]string = make(map[string]string)

	for _, param := range parameters {
		if _, ok := m[param.Name]; ok {
			if param.Conditions != nil {
				for _, con := range param.Conditions {
					if con.Params != nil {
						for _, p := range con.Params {
							m[param.Name] = fmt.Sprintf("%v <br> %v=%v:%v ", m[param.Name], p.Name, p.Value, param.Default)
						}
					}
				}
			}

		} else {
			if param.Conditions != nil {
				m[param.Name] = ""
				for _, con := range param.Conditions {
					if con.Params != nil {
						for _, p := range con.Params {
							m[param.Name] += fmt.Sprintf("%v=%v:%v", p.Name, p.Value, param.Default)
						}
					}
				}
			} else {
				m[param.Name] = fmt.Sprintf("%v", param.Default)
			}
		}
	}

	return m
}

func createConfigurationTable(parameters []config.StepParameters) string {

	var table = "| parameter | general | step/stage |\n"
	table += "|-----------|---------|------------|\n"

	for _, param := range parameters {
		if len(param.Scope) > 0 {
			general := contains(param.Scope, "GENERAL")
			step := contains(param.Scope, "STEPS")

			table += fmt.Sprintf(" | %v | %v | %v | \n ", param.Name, ifThenElse(general, "X", ""), ifThenElse(step, "X", ""))
		}
	}

	return table
}

func handleStepParameters(stepData *config.StepData) {
	//add secrets to pstep arameters
	appendSecretsToParameters(stepData)

	//get the context defaults
	context := getDocuContextDefaults(stepData)
	if len(context) > 0 {
		contextDefaultPath := "pkg/generator/helper/piper-context-defaults.yaml"
		mCD := readContextDefaultDescription(contextDefaultPath)
		//create StepParemeters items for context defaults
		for k, v := range context {
			if len(v) > 0 {
				stepData.Spec.Inputs.Parameters = append(stepData.Spec.Inputs.Parameters, config.StepParameters{Name: k, Default: v, Mandatory: false, Description: mCD[k]})
			}
		}
	}
	//Sort Parameters
	sortStepParameters(stepData)
}

func appendSecretsToParameters(stepData *config.StepData) {

	secrets := stepData.Spec.Inputs.Secrets
	if secrets != nil {
		for _, secret := range secrets {
			item := config.StepParameters{Name: secret.Name, Type: secret.Type, Description: secret.Description, Mandatory: true}
			stepData.Spec.Inputs.Parameters = append(stepData.Spec.Inputs.Parameters, item)
		}
	}
}

func getDocuContextDefaults(m *config.StepData) map[string]string {

	var result map[string]string = make(map[string]string)

	//creates the context defaults for containers
	if len(m.Spec.Containers) > 0 {
		keys := []string{}
		resources := map[string][]string{}
		for _, container := range m.Spec.Containers {
			key := ""
			if len(container.Conditions) > 0 {
				key = fmt.Sprintf("%v=%v", container.Conditions[0].Params[0].Name, container.Conditions[0].Params[0].Value)
			}
			if len(container.Command) > 0 {
				keys = append(keys, key+"_containerCommand")
			}
			if m.Metadata.Name == "dockerExecuteOnKubernetes" {
				keys = append(keys, key+"_containerName")
			}
			keys = append(keys, key+"_containerShell")
			keys = append(keys, key+"_dockerEnvVars")
			keys = append(keys, key+"_dockerImage")
			keys = append(keys, key+"_dockerName")
			keys = append(keys, key+"_dockerPullImage")
			keys = append(keys, key+"_dockerWorkspace")

			workingDir := ifThenElse(len(container.WorkingDir) > 0, container.WorkingDir, "\\<empty\\>")
			resources[key+"_containerShell"] = append(resources[key+"_containerShell"], container.Shell)
			resources[key+"_dockerName"] = append(resources[key+"_dockerName"], container.Name)

			//Only for Step: dockerExecuteOnKubernetes
			if m.Metadata.Name == "dockerExecuteOnKubernetes" {
				resources[key+"_containerName"] = append(resources[key+"_containerName"], container.Name)
			}
			//ContainerCommand > 0
			if len(container.Command) > 0 {
				resources[key+"_containerCommand"] = append(resources[key+"_containerCommand"], container.Command[0])
			}
			//ImagePullPolicy > 0
			if len(container.ImagePullPolicy) > 0 {
				resources[key+"_dockerPullImage"] = []string{fmt.Sprintf("%v", container.ImagePullPolicy != "Never")}
			}

			//Different when key is set (Param.Name + Param.Value)
			if len(key) > 0 {
				resources[key+"_dockerEnvVars"] = append(resources[key+"_dockerEnvVars"], fmt.Sprintf("%v:\\[%v\\]", key, strings.Join(envVarsAsStringSlice(container.EnvVars), "")))
				resources[key+"_dockerImage"] = append(resources[key+"_dockerImage"], fmt.Sprintf("%v:%v", key, container.Image))
				resources[key+"_dockerWorkspace"] = append(resources[key+"_dockerWorkspace"], fmt.Sprintf("%v:%v", key, workingDir))
			} else {
				resources[key+"_dockerEnvVars"] = append(resources[key+"_dockerEnvVars"], fmt.Sprintf("%v", strings.Join(envVarsAsStringSlice(container.EnvVars), "")))
				resources[key+"_dockerImage"] = append(resources[key+"_dockerImage"], container.Image)
				resources[key+"_dockerWorkspace"] = append(resources[key+"_dockerWorkspace"], workingDir)
			}
			// Ready command not relevant for main runtime container so far
			//p[] = container.ReadyCommand
		}

		for _, key := range keys {
			s := strings.Split(key, "_")
			if len(strings.Join(resources[key], ", ")) > 1 {
				result[s[1]] += fmt.Sprintf("%v <br>", strings.Join(resources[key], ", "))
			} else if len(strings.Join(resources[key], ", ")) == 1 {
				if _, ok := result[s[1]]; !ok {
					result[s[1]] = fmt.Sprintf("%v", strings.Join(resources[key], ", "))
				}
			}
		}
	}

	//creates the context defaults for sidecars
	if len(m.Spec.Sidecars) > 0 {
		if len(m.Spec.Sidecars[0].Command) > 0 {
			result["sidecarCommand"] += m.Spec.Sidecars[0].Command[0]
		}
		result["sidecarEnvVars"] = strings.Join(envVarsAsStringSlice(m.Spec.Sidecars[0].EnvVars), "")
		result["sidecarImage"] = m.Spec.Sidecars[0].Image
		result["sidecarName"] = m.Spec.Sidecars[0].Name
		if len(m.Spec.Sidecars[0].ImagePullPolicy) > 0 {
			result["sidecarPullImage"] = fmt.Sprintf("%v", m.Spec.Sidecars[0].ImagePullPolicy != "Never")
		}
		result["sidecarReadyCommand"] = m.Spec.Sidecars[0].ReadyCommand
		result["sidecarWorkspace"] = m.Spec.Sidecars[0].WorkingDir
	}

	// not filled for now since this is not relevant in Kubernetes case
	//p["dockerOptions"] = container.
	//p["dockerVolumeBind"] = container.
	//root["containerPortMappings"] = m.Spec.Sidecars[0].
	//root["sidecarOptions"] = m.Spec.Sidecars[0].
	//root["sidecarVolumeBind"] = m.Spec.Sidecars[0].

	//creates the context defaults for resources
	if len(m.Spec.Inputs.Resources) > 0 {
		keys := []string{}
		resources := map[string][]string{}

		//fill the map with the key (condition) and the values (resource.Name) to combine the conditions under the resource.Name
		for _, resource := range m.Spec.Inputs.Resources {
			if resource.Type == "stash" {
				key := ""
				if len(resource.Conditions) > 0 {
					key = fmt.Sprintf("%v=%v", resource.Conditions[0].Params[0].Name, resource.Conditions[0].Params[0].Value)
				}
				if resources[key] == nil {
					keys = append(keys, key)
					resources[key] = []string{}
				}
				resources[key] = append(resources[key], resource.Name)
			}
		}

		for _, key := range keys {
			//more than one key when there are conditions
			if len(key) > 0 {
				result["stashContent"] += fmt.Sprintf("%v:\\[%v\\] <br>", key, strings.Join(resources[key], ", "))
			} else {
				//single entry for stash content (no condition)
				result["stashContent"] += fmt.Sprintf("\\[%v\\] <br>", strings.Join(resources[key], ", "))
			}
		}
	}

	return result
}

func envVarsAsStringSlice(envVars []config.EnvVar) []string {
	e := []string{}
	c := len(envVars) - 1
	for k, v := range envVars {
		if k < c {
			e = append(e, fmt.Sprintf("%v=%v, <br>", v.Name, ifThenElse(len(v.Value) > 0, v.Value, "\\<empty\\>")))
		} else {
			e = append(e, fmt.Sprintf("%v=%v", v.Name, ifThenElse(len(v.Value) > 0, v.Value, "\\<empty\\>")))
		}
	}
	return e
}

func sortStepParameters(stepData *config.StepData) {

	if stepData.Spec.Inputs.Parameters != nil {
		parameters := stepData.Spec.Inputs.Parameters

		sort.Slice(parameters[:], func(i, j int) bool {
			return parameters[i].Name < parameters[j].Name
		})
	}
}
