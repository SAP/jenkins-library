package helper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/config"
)

// generates the step documentation and replaces the template with the generated documentation
func generateStepDocumentation(stepData config.StepData, docuHelperData DocuHelperData) error {
	fmt.Printf("Generate docu for: %v\n", stepData.Metadata.Name)
	//create the file path for the template and open it.
	docTemplateFilePath := fmt.Sprintf("%v%v.md", docuHelperData.DocTemplatePath, stepData.Metadata.Name)
	docTemplate, err := docuHelperData.OpenDocTemplateFile(docTemplateFilePath)

	if docTemplate != nil {
		defer docTemplate.Close()
	}
	//check if there is an error during opening the template (true : skip docu generation for this meta data file)
	if err != nil {
		return fmt.Errorf("error occured: %v", err)
	}

	content := readAndAdjustTemplate(docTemplate)
	if len(content) <= 0 {
		return fmt.Errorf("error occured: no content inside of the template")
	}

	// binding of functions and placeholder
	funcMap := template.FuncMap{
		"docGenStepName":      docGenStepName,
		"docGenDescription":   docGenDescription,
		"docGenParameters":    docGenParameters,
		"docGenConfiguration": docGenConfiguration,
	}
	tmpl, err := template.New("doc").Funcs(funcMap).Parse(content)
	checkError(err)

	setDefaultStepParameters(&stepData)
	// add secrets, context defaults to the step parameters
	handleStepParameters(&stepData)

	//write executed template data to the previously opened file
	var docContent bytes.Buffer
	err = tmpl.Execute(&docContent, stepData)
	checkError(err)

	//overwrite existing file
	err = docuHelperData.DocFileWriter(docTemplateFilePath, docContent.Bytes(), 644)
	checkError(err)

	fmt.Printf("Documentation generation complete for: %v\n", stepData.Metadata.Name)

	return nil
}

func setDefaultStepParameters(stepData *config.StepData) {
	for k, param := range stepData.Spec.Inputs.Parameters {
		if param.Default == nil {
			switch param.Type {
			case "bool":
				param.Default = "`false`"
			case "int":
				param.Default = "`0`"
			}
		} else {
			switch param.Type {
			case "[]string":
				param.Default = fmt.Sprintf("`%v`", param.Default)
			case "string":
				param.Default = fmt.Sprintf("`%v`", param.Default)
			case "bool":
				param.Default = fmt.Sprintf("`%v`", param.Default)
			case "int":
				param.Default = fmt.Sprintf("`%v`", param.Default)
			}
		}
		stepData.Spec.Inputs.Parameters[k] = param
	}
}

func readAndAdjustTemplate(docFile io.ReadCloser) string {
	//read template content
	content, err := ioutil.ReadAll(docFile)
	checkError(err)
	contentStr := string(content)

	//replace old placeholder with new ones
	contentStr = strings.ReplaceAll(contentStr, "${docGenStepName}", "{{docGenStepName .}}")
	contentStr = strings.ReplaceAll(contentStr, "${docGenDescription}", "{{docGenDescription .}}")
	contentStr = strings.ReplaceAll(contentStr, "${docGenParameters}", "{{docGenParameters .}}")
	contentStr = strings.ReplaceAll(contentStr, "${docGenConfiguration}", "{{docGenConfiguration .}}")
	contentStr = strings.ReplaceAll(contentStr, "## ${docJenkinsPluginDependencies}", "")

	return contentStr
}

// Replaces the docGenStepName placeholder with the content from the yaml
func docGenStepName(stepData config.StepData) string {
	return stepData.Metadata.Name
}

// Replaces the docGenDescription placeholder with content from the yaml
func docGenDescription(stepData config.StepData) string {
	return "Description\n\n" + stepData.Metadata.LongDescription
}

// Replaces the docGenParameters placeholder with the content from the yaml
func docGenParameters(stepData config.StepData) string {
	var parameters = ""
	//create step parameter table
	parameters += createParametersTable(stepData.Spec.Inputs.Parameters) + "\n"
	//create parameters detail section
	parameters += createParametersDetail(stepData.Spec.Inputs.Parameters)
	return "Parameters\n\n" + parameters
}

// Replaces the docGenConfiguration placeholder with the content from the yaml
func docGenConfiguration(stepData config.StepData) string {
	var configuration = "We recommend to define values of step parameters via [config.yml file](../configuration.md).\n\n"
	configuration += "In following sections of the config.yml the configuration is possible:\n\n"
	// create step configuration table
	configuration += createConfigurationTable(stepData.Spec.Inputs.Parameters)
	return "Step Configuration\n\n" + configuration
}

func createParametersTable(parameters []config.StepParameters) string {
	var table = "| name | mandatory | default | possible values |\n"
	table += "| ---- | --------- | ------- | --------------- |\n"

	m := combineEqualParametersTogether(parameters)

	for _, param := range parameters {
		if v, ok := m[param.Name]; ok {
			table += fmt.Sprintf("| `%v` | %v | %v | %v |\n", param.Name, ifThenElse(param.Mandatory && param.Default == nil, "Yes", "No"), ifThenElse(v == "<nil>", "", v), possibleValuesToString(param.PossibleValues))
			delete(m, param.Name)
		}
	}
	return table
}

func createParametersDetail(parameters []config.StepParameters) string {
	var details = ""
	var m map[string]bool = make(map[string]bool)
	for _, param := range parameters {
		if _, ok := m[param.Name]; !ok {
			if len(param.Description) > 0 {
				details += fmt.Sprintf(" * `%v`: %v\n", param.Name, param.Description)
				m[param.Name] = true
			}
		}
	}
	return details
}

//combines equal parameters and the values
func combineEqualParametersTogether(parameters []config.StepParameters) map[string]string {
	var m map[string]string = make(map[string]string)

	for _, param := range parameters {
		m[param.Name] = fmt.Sprintf("%v", param.Default)

		if _, ok := m[param.Name]; ok {
			addExistingParameterWithCondition(param, m)
		} else {
			addNewParameterWithCondition(param, m)
		}
	}
	return m
}

func addExistingParameterWithCondition(param config.StepParameters, m map[string]string) {
	if param.Conditions != nil {
		for _, con := range param.Conditions {
			if con.Params != nil {
				for _, p := range con.Params {
					m[param.Name] = fmt.Sprintf("%v<br>%v=`%v`: `%v` ", m[param.Name], p.Name, p.Value, param.Default)
				}
			}
		}
	}
}

func addNewParameterWithCondition(param config.StepParameters, m map[string]string) {
	if param.Conditions != nil {
		m[param.Name] = ""
		for _, con := range param.Conditions {
			if con.Params != nil {
				for _, p := range con.Params {
					m[param.Name] += fmt.Sprintf("%v=`%v`: `%v` ", p.Name, p.Value, param.Default)
				}
			}
		}
	}
}

func createConfigurationTable(parameters []config.StepParameters) string {
	var table = "| parameter | general | step/stage |\n"
	table += "| --------- | ------- | ---------- |\n"

	for _, param := range parameters {
		if len(param.Scope) > 0 {
			general := contains(param.Scope, "GENERAL")
			step := contains(param.Scope, "STEPS")

			table += fmt.Sprintf("| `%v` | %v | %v |\n", param.Name, ifThenElse(general, "X", ""), ifThenElse(step, "X", ""))
		}
	}
	return table
}

func handleStepParameters(stepData *config.StepData) {
	//add secrets to step parameters
	appendSecretsToParameters(stepData)

	//get the context defaults
	context := getDocuContextDefaults(stepData)
	if len(context) > 0 {
		contextDefaultPath := "pkg/generator/helper/piper-context-defaults.yaml"
		mCD := readContextDefaultDescription(contextDefaultPath)
		//fmt.Printf("ContextDefault Map: %v \n", context)
		//create StepParemeters items for context defaults
		for k, v := range context {
			if len(v) > 0 {
				//containerName only for Step: dockerExecuteOnKubernetes
				if k != "containerName" || stepData.Metadata.Name == "dockerExecuteOnKubernetes" {
					if mCD[k] != nil {
						cdp := mCD[k].(ContextDefaultParameters)
						stepData.Spec.Inputs.Parameters = append(stepData.Spec.Inputs.Parameters, config.StepParameters{Name: k, Default: v, Mandatory: false, Description: cdp.Description, Scope: cdp.Scope})
					}
				}
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

func getDocuContextDefaults(step *config.StepData) map[string]string {
	var result map[string]string = make(map[string]string)

	//creates the context defaults for containers
	addDefaultContainerContent(step, result)
	//creates the context defaults for sidecars
	addDefaultSidecarContent(step, result)
	//creates the context defaults for resources
	addStashContent(step, result)

	return result
}

func addDefaultContainerContent(m *config.StepData, result map[string]string) {
	//creates the context defaults for containers
	if len(m.Spec.Containers) > 0 {
		keys := map[string][]string{}
		resources := map[string][]string{}
		bEmptyKey := true
		for _, container := range m.Spec.Containers {

			addContainerValues(container, bEmptyKey, resources, keys)
		}
		createDefaultContainerEntries(keys, resources, result)
	}
}

func addContainerValues(container config.Container, bEmptyKey bool, resources map[string][]string, m map[string][]string) {
	//create keys
	key := ""
	if len(container.Conditions) > 0 {
		key = fmt.Sprintf("%v=`%v`", container.Conditions[0].Params[0].Name, container.Conditions[0].Params[0].Value)
	}

	//only add the key ones
	if bEmptyKey || len(key) > 0 {

		if len(container.Command) > 0 {
			m["containerCommand"] = append(m["containerCommand"], key+"_containerCommand")
		}
		m["containerName"] = append(m["containerName"], key+"_containerName")
		m["containerShell"] = append(m["containerShell"], key+"_containerShell")
		m["dockerEnvVars"] = append(m["dockerEnvVars"], key+"_dockerEnvVars")
		m["dockerImage"] = append(m["dockerImage"], key+"_dockerImage")
		m["dockerName"] = append(m["dockerName"], key+"_dockerName")
		m["dockerPullImage"] = append(m["dockerPullImage"], key+"_dockerPullImage")
		m["dockerOptions"] = append(m["dockerOptions"], key+"_dockerOptions")
		m["dockerWorkspace"] = append(m["dockerWorkspace"], key+"_dockerWorkspace")
	}

	if len(container.Conditions) == 0 {
		bEmptyKey = false
	}

	//add values
	addValuesToMap(container, key, resources)
}

func addValuesToMap(container config.Container, key string, resources map[string][]string) {
	if len(container.Name) > 0 {
		resources[key+"_containerName"] = append(resources[key+"_containerName"], "`"+container.Name+"`")
	}
	//ContainerShell > 0
	if len(container.Shell) > 0 {
		resources[key+"_containerShell"] = append(resources[key+"_containerShell"], "`"+container.Shell+"`")
	}
	if len(container.Name) > 0 {
		resources[key+"_dockerName"] = append(resources[key+"_dockerName"], "`"+container.Name+"`")
	}

	//ContainerCommand > 0
	if len(container.Command) > 0 {
		resources[key+"_containerCommand"] = append(resources[key+"_containerCommand"], "`"+container.Command[0]+"`")
	}
	//ImagePullPolicy > 0
	if len(container.ImagePullPolicy) > 0 {
		resources[key+"_dockerPullImage"] = []string{fmt.Sprintf("`%v`", container.ImagePullPolicy != "Never")}
	}
	//Different when key is set (Param.Name + Param.Value)
	workingDir := ifThenElse(len(container.WorkingDir) > 0, "`"+container.WorkingDir+"`", "\\<empty\\>")
	if len(key) > 0 {
		resources[key+"_dockerEnvVars"] = append(resources[key+"_dockerEnvVars"], fmt.Sprintf("%v: `[%v]`", key, strings.Join(envVarsAsStringSlice(container.EnvVars), "")))
		resources[key+"_dockerImage"] = append(resources[key+"_dockerImage"], fmt.Sprintf("%v: `%v`", key, container.Image))
		resources[key+"_dockerOptions"] = append(resources[key+"_dockerOptions"], fmt.Sprintf("%v: `[%v]`", key, strings.Join(optionsAsStringSlice(container.Options), "")))
		resources[key+"_dockerWorkspace"] = append(resources[key+"_dockerWorkspace"], fmt.Sprintf("%v: %v", key, workingDir))
	} else {
		resources[key+"_dockerEnvVars"] = append(resources[key+"_dockerEnvVars"], fmt.Sprintf("`[%v]`", strings.Join(envVarsAsStringSlice(container.EnvVars), "")))
		resources[key+"_dockerImage"] = append(resources[key+"_dockerImage"], "`"+container.Image+"`")
		resources[key+"_dockerOptions"] = append(resources[key+"_dockerOptions"], fmt.Sprintf("`[%v]`", strings.Join(optionsAsStringSlice(container.Options), "")))
		resources[key+"_dockerWorkspace"] = append(resources[key+"_dockerWorkspace"], workingDir)
	}
}

func createDefaultContainerEntries(keys map[string][]string, resources map[string][]string, result map[string]string) {
	//loop over keys map, key is the description of the parameter for example : dockerEnvVars, ...
	for k, p := range keys {
		if p != nil {
			//loop over key array to get the values from the resources
			for _, key := range p {
				doLineBreak := !strings.HasPrefix(key, "_")

				if len(strings.Join(resources[key], ", ")) > 1 {
					result[k] += fmt.Sprintf("%v", strings.Join(resources[key], ", "))
					if doLineBreak {
						result[k] += "<br>"
					}
				} else if len(strings.Join(resources[key], ", ")) == 1 {
					if _, ok := result[k]; !ok {
						result[k] = fmt.Sprintf("%v", strings.Join(resources[key], ", "))
					} else {
						result[k] += fmt.Sprintf("%v", strings.Join(resources[key], ", "))
						if doLineBreak {
							result[k] += "<br>"
						}
					}
				}
			}
		}
	}
}

func addDefaultSidecarContent(m *config.StepData, result map[string]string) {
	//creates the context defaults for sidecars
	if len(m.Spec.Sidecars) > 0 {
		if len(m.Spec.Sidecars[0].Command) > 0 {
			result["sidecarCommand"] += m.Spec.Sidecars[0].Command[0]
		}
		result["sidecarEnvVars"] = strings.Join(envVarsAsStringSlice(m.Spec.Sidecars[0].EnvVars), "")
		result["sidecarImage"] = fmt.Sprintf("`%s`", m.Spec.Sidecars[0].Image)
		result["sidecarName"] = fmt.Sprintf("`%s`", m.Spec.Sidecars[0].Name)
		if len(m.Spec.Sidecars[0].ImagePullPolicy) > 0 {
			result["sidecarPullImage"] = fmt.Sprintf("%v", m.Spec.Sidecars[0].ImagePullPolicy != "Never")
		}
		result["sidecarReadyCommand"] = m.Spec.Sidecars[0].ReadyCommand
		result["sidecarOptions"] = strings.Join(optionsAsStringSlice(m.Spec.Sidecars[0].Options), "")
		result["sidecarWorkspace"] = m.Spec.Sidecars[0].WorkingDir
	}
}

func addStashContent(m *config.StepData, result map[string]string) {
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
				result["stashContent"] += fmt.Sprintf("%v: `[%v]` <br>", key, strings.Join(resources[key], ", "))
			} else {
				//single entry for stash content (no condition)
				result["stashContent"] += fmt.Sprintf("`[%v]`", strings.Join(resources[key], ", "))
			}
		}
	}
}

func envVarsAsStringSlice(envVars []config.EnvVar) []string {
	e := []string{}
	c := len(envVars) - 1
	for k, v := range envVars {
		if k < c {
			e = append(e, fmt.Sprintf("%v=%v ", v.Name, ifThenElse(len(v.Value) > 0, v.Value, "\\<empty\\>")))
		} else {
			e = append(e, fmt.Sprintf("%v=%v", v.Name, ifThenElse(len(v.Value) > 0, v.Value, "\\<empty\\>")))
		}
	}
	return e
}

func optionsAsStringSlice(options []config.Option) []string {
	e := []string{}
	c := len(options) - 1
	for k, v := range options {
		if k < c {
			e = append(e, fmt.Sprintf("%v %v ", v.Name, ifThenElse(len(v.Value) > 0, v.Value, "\\<empty\\>")))
		} else {
			e = append(e, fmt.Sprintf("%v %v", v.Name, ifThenElse(len(v.Value) > 0, v.Value, "\\<empty\\>")))
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

func possibleValuesToString(in []interface{}) (out string) {
	if len(in) == 0 {
		return
	}
	out = fmt.Sprintf("`%v`", in[0])
	if len(in) == 1 {
		return
	}
	for _, value := range in[1:] {
		out += fmt.Sprintf(", `%v`", value)
	}
	return
}
