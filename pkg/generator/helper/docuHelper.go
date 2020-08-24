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

var stepParameterNames []string

// generates the step documentation and replaces the template with the generated documentation
func generateStepDocumentation(stepData config.StepData, docuHelperData DocuHelperData) error {
	fmt.Printf("Generate docu for: %v\n", stepData.Metadata.Name)
	//create the file path for the template and open it.
	docTemplateFilePath := fmt.Sprintf("%v%v.md", docuHelperData.DocTemplatePath, stepData.Metadata.Name)
	docTemplate, err := docuHelperData.OpenDocTemplateFile(docTemplateFilePath)

	if docTemplate != nil {
		defer docTemplate.Close()
	}
	// check if there is an error during opening the template (true : skip docu generation for this meta data file)
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

	// add secrets, context defaults to the step parameters
	handleStepParameters(&stepData)

	// write executed template data to the previously opened file
	var docContent bytes.Buffer
	err = tmpl.Execute(&docContent, &stepData)
	checkError(err)

	// overwrite existing file
	err = docuHelperData.DocFileWriter(docTemplateFilePath, docContent.Bytes(), 644)
	checkError(err)

	fmt.Printf("Documentation generation complete for: %v\n", stepData.Metadata.Name)

	return nil
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
func docGenStepName(stepData *config.StepData) string {
	return stepData.Metadata.Name + "\n\n" + stepData.Metadata.Description + "\n"
}

// Replaces the docGenDescription placeholder with content from the yaml
func docGenDescription(stepData *config.StepData) string {
	description := ""

	description += "Description\n\n" + stepData.Metadata.LongDescription + "\n\n"

	description += "## Usage\n\n"
	description += "We recommend to define values of [step parameters](#parameters) via [config.yml file](../configuration.md). In this case, calling the step is reduced to one simple line.<br />Calling the step can be done either via the Jenkins library step or on the [command line](../cli/index.md).\n\n"
	description += "### Jenkins pipelines\n\n```groovy\n"
	description += fmt.Sprintf("%v script: this\n```\n", stepData.Metadata.Name)
	description += "### Command line\n\n```\n"
	description += fmt.Sprintf("piper %v\n```\n\n", stepData.Metadata.Name)
	description += stepOutputs(stepData)
	return description
}

func stepOutputs(stepData *config.StepData) string {
	if len(stepData.Spec.Outputs.Resources) == 0 {
		return ""
	}

	stepOutput := "\n## Outputs\n\n"
	stepOutput += "| Output type | Details |\n"
	stepOutput += "| ----------- | ------- |\n"

	for _, res := range stepData.Spec.Outputs.Resources {
		//handle commonPipelineEnvironment output
		if res.Type == "piperEnvironment" {
			stepOutput += fmt.Sprintf("| %v | <ul>", res.Name)
			for _, param := range res.Parameters {
				stepOutput += fmt.Sprintf("<li>%v</li>", param["name"])
			}
			stepOutput += "</ul> |\n"
		}

		//handle Influx output
		if res.Type == "influx" {
			stepOutput += fmt.Sprintf("| %v | ", res.Name)
			for _, param := range res.Parameters {
				stepOutput += fmt.Sprintf("measurement `%v`<br /><ul>", param["name"])
				fields, _ := param["fields"].([]interface{})
				for _, field := range fields {
					fieldMap, _ := field.(map[string]interface{})
					stepOutput += fmt.Sprintf("<li>%v</li>", fieldMap["name"])
				}
			}
			stepOutput += "</ul> |\n"
		}

	}

	return stepOutput

}

// Replaces the docGenParameters placeholder with the content from the yaml
func docGenParameters(stepData *config.StepData) string {

	var parameters = "Parameters\n\n"

	// sort parameters alphabetically with mandatory parameters first
	sortStepParameters(stepData, true)
	parameters += "### Overview\n\n"
	parameters += createParameterOverview(stepData)

	// sort parameters alphabetically
	sortStepParameters(stepData, false)
	parameters += "### Details\n\n"
	parameters += createParameterDetails(stepData)

	return parameters
}

// Replaces the docGenConfiguration placeholder with the content from the yaml
func docGenConfiguration(stepData *config.StepData) string {
	//not used anymore -> part of Parameters section
	return ""
}

func createParameterOverview(stepData *config.StepData) string {
	var table = "| Name | Mandatory | Additional information |\n"
	table += "| ---- | --------- | ---------------------- |\n"

	for _, param := range stepData.Spec.Inputs.Parameters {
		table += fmt.Sprintf("| [%v](#%v) | %v | %v |\n", param.Name, strings.ToLower(param.Name), ifThenElse(param.Mandatory, "**yes**", "no"), parameterFurtherInfo(param.Name, stepData))
	}

	table += "\n"

	return table
}

func parameterFurtherInfo(paramName string, stepData *config.StepData) string {

	// handle general parameters
	// ToDo: add special handling once we have more than one general parameter to consider
	if paramName == "verbose" {
		return "activates debug output"
	}

	if paramName == "script" {
		return "[![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#) reference to Jenkins main pipeline script"
	}

	// handle Jenkins-specific parameters
	if !contains(stepParameterNames, paramName) {
		for _, secret := range stepData.Spec.Inputs.Secrets {
			if paramName == secret.Name && secret.Type == "jenkins" {
				return "[![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#) id of credentials ([using credentials](https://www.jenkins.io/doc/book/using/using-credentials/))"
			}
		}
		return "[![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#)"
	}

	// handle Secrets
	for _, param := range stepData.Spec.Inputs.Parameters {
		if paramName == param.Name {
			if param.Secret {
				secretInfo := "[![Secret](https://img.shields.io/badge/-Secret-yellowgreen)](#) pass via ENV or Jenkins credentials"
				for _, res := range param.ResourceRef {
					if res.Type == "secret" {
						secretInfo += fmt.Sprintf(" ([`%v`](#%v))", res.Name, strings.ToLower(res.Name))
					}
				}
				return secretInfo
			}
			return ""
		}
	}
	return ""
}

func createParameterDetails(stepData *config.StepData) string {

	details := ""

	//jenkinsParameters := append(jenkinsParameters(stepData), "script")

	for _, param := range stepData.Spec.Inputs.Parameters {
		details += fmt.Sprintf("#### %v\n\n", param.Name)

		if !contains(stepParameterNames, param.Name) {
			details += "**Jenkins-specific:** Used for proper environment setup.\n\n"
		}

		if len(param.LongDescription) > 0 {
			details += param.LongDescription + "\n\n"
		} else {
			details += param.Description + "\n\n"
		}

		details += "[back to overview](#parameters)\n\n"

		details += "| Scope | Details |\n"
		details += "| ---- | --------- |\n"

		details += fmt.Sprintf("| Aliases | %v |\n", aliasList(param.Aliases))
		details += fmt.Sprintf("| Type | `%v` |\n", param.Type)
		details += fmt.Sprintf("| Mandatory | %v |\n", ifThenElse(param.Mandatory && param.Default == nil, "**yes**", "no"))
		details += fmt.Sprintf("| Default | %v |\n", formatDefault(param, stepParameterNames))
		if param.PossibleValues != nil {
			details += fmt.Sprintf("| Possible values | %v |\n", possibleValueList(param.PossibleValues))
		}
		details += fmt.Sprintf("| Secret | %v |\n", ifThenElse(param.Secret, "**yes**", "no"))
		details += fmt.Sprintf("| Configuration scope | %v |\n", scopeDetails(param.Scope))
		details += fmt.Sprintf("| Resource references | %v |\n", resourceReferenceDetails(param.ResourceRef))

		details += "\n\n"
	}

	return details
}

func formatDefault(param config.StepParameters, stepParameterNames []string) string {
	if param.Default == nil {
		// Return environment variable for all step parameters (not for Jenkins-specific parameters) in case no default is available
		if contains(stepParameterNames, param.Name) {
			return fmt.Sprintf("`$PIPER_%v` (if set)", param.Name)
		}
		return ""
	}
	//first consider conditional defaults
	switch v := param.Default.(type) {
	case []conditionDefault:
		defaults := []string{}
		for _, condDef := range v {
			//ToDo: add type-specific handling of default
			defaults = append(defaults, fmt.Sprintf("%v=`%v`: `%v`", condDef.key, condDef.value, condDef.def))
		}
		return strings.Join(defaults, "<br />")
	case []interface{}:
		// handle for example stashes which possibly contain a mixture of fix and conditional values
		defaults := []string{}
		for _, def := range v {
			if condDef, ok := def.(conditionDefault); ok {
				defaults = append(defaults, fmt.Sprintf("%v=`%v`: `%v`", condDef.key, condDef.value, condDef.def))
			} else {
				defaults = append(defaults, fmt.Sprintf("- `%v`", def))
			}
		}
		return strings.Join(defaults, "<br />")
	case map[string]string:
		defaults := []string{}
		for key, def := range v {
			defaults = append(defaults, fmt.Sprintf("`%v`: `%v`", key, def))
		}
		return strings.Join(defaults, "<br />")
	case string:
		if len(v) == 0 {
			return "`''`"
		}
		return fmt.Sprintf("`%v`", v)
	default:
		return fmt.Sprintf("`%v`", param.Default)
	}
}

func aliasList(aliases []config.Alias) string {
	switch len(aliases) {
	case 0:
		return "-"
	case 1:
		alias := fmt.Sprintf("`%v`", aliases[0].Name)
		if aliases[0].Deprecated {
			alias += " (**deprecated**)"
		}
		return alias
	default:
		aList := make([]string, len(aliases))
		for i, alias := range aliases {
			aList[i] = fmt.Sprintf("- `%v`", alias.Name)
			if alias.Deprecated {
				aList[i] += " (**deprecated**)"
			}
		}
		return strings.Join(aList, "<br />")
	}
}

func possibleValueList(possibleValues []interface{}) string {
	if len(possibleValues) == 0 {
		return ""
	}

	pList := make([]string, len(possibleValues))
	for i, possibleValue := range possibleValues {
		pList[i] = fmt.Sprintf("- `%v`", fmt.Sprint(possibleValue))
	}
	return strings.Join(pList, "<br />")
}

func scopeDetails(scope []string) string {
	scopeDetails := "<ul>"
	scopeDetails += fmt.Sprintf("<li>%v parameter</li>", ifThenElse(contains(scope, "PARAMETERS"), "&#9746;", "&#9744;"))
	scopeDetails += fmt.Sprintf("<li>%v general</li>", ifThenElse(contains(scope, "GENERAL"), "&#9746;", "&#9744;"))
	scopeDetails += fmt.Sprintf("<li>%v steps</li>", ifThenElse(contains(scope, "STEPS"), "&#9746;", "&#9744;"))
	scopeDetails += fmt.Sprintf("<li>%v stages</li>", ifThenElse(contains(scope, "STAGES"), "&#9746;", "&#9744;"))
	scopeDetails += "</ul>"
	return scopeDetails
}

func resourceReferenceDetails(resourceRef []config.ResourceReference) string {

	if len(resourceRef) == 0 {
		return "none"
	}

	resourceDetails := ""
	for _, resource := range resourceRef {
		if resource.Name == "commonPipelineEnvironment" {
			resourceDetails += "_commonPipelineEnvironment_:<br />"
			resourceDetails += fmt.Sprintf("&nbsp;&nbsp;reference to: `%v`<br />", resource.Param)
			continue
		}

		if resource.Type == "secret" {
			resourceDetails += "Jenkins credential id:<br />"
			for i, alias := range resource.Aliases {
				if i == 0 {
					resourceDetails += "&nbsp;&nbsp;aliases:<br />"
				}
				resourceDetails += fmt.Sprintf("&nbsp;&nbsp;- `%v`%v<br />", alias.Name, ifThenElse(alias.Deprecated, " (**Deprecated**)", ""))
			}
			resourceDetails += fmt.Sprintf("&nbsp;&nbsp;id: `%v`<br />", resource.Name)
			resourceDetails += fmt.Sprintf("&nbsp;&nbsp;reference to: `%v`<br />", resource.Param)
			continue
		}
	}

	return resourceDetails
}

func handleStepParameters(stepData *config.StepData) {

	stepParameterNames = stepData.GetParameterFilters().All

	//add general options like script, verbose, etc.
	//ToDo: add to context.yaml
	appendGeneralOptionsToParameters(stepData)

	//add secrets to step parameters
	appendSecretsToParameters(stepData)

	//consolidate conditional parameters:
	//- remove duplicate parameter entries
	//- combine defaults (consider conditions)
	consolidateConditionalParameters(stepData)

	//get the context defaults
	appendContextParameters(stepData)

	//consolidate context defaults:
	//- combine defaults (consider conditions)
	consolidateContextDefaults(stepData)

	setDefaultAndPossisbleValues(stepData)
}

func appendGeneralOptionsToParameters(stepData *config.StepData) {
	script := config.StepParameters{
		Name: "script", Type: "Jenkins Script", Mandatory: true,
		Description: "The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` for retrieving, e.g. configuration parameters.",
	}
	verbose := config.StepParameters{
		Name: "verbose", Type: "bool", Mandatory: false, Default: false, Scope: []string{"PARAMETERS", "GENERAL", "STEPS", "STAGES"},
		Description: "verbose output",
	}
	stepData.Spec.Inputs.Parameters = append(stepData.Spec.Inputs.Parameters, script, verbose)
}

func appendSecretsToParameters(stepData *config.StepData) {
	secrets := stepData.Spec.Inputs.Secrets
	if secrets != nil {
		for _, secret := range secrets {
			item := config.StepParameters{Name: secret.Name, Type: "string", Scope: []string{"PARAMETERS", "GENERAL", "STEPS", "STAGES"}, Description: secret.Description, Mandatory: true}
			stepData.Spec.Inputs.Parameters = append(stepData.Spec.Inputs.Parameters, item)
		}
	}
}

type paramConditionDefaults map[string]*conditionDefaults

type conditionDefaults struct {
	equal []conditionDefault
}

type conditionDefault struct {
	key   string
	value string
	def   interface{}
}

func consolidateConditionalParameters(stepData *config.StepData) {
	newParamList := []config.StepParameters{}

	paramConditions := paramConditionDefaults{}

	for _, param := range stepData.Spec.Inputs.Parameters {
		if param.Conditions == nil || len(param.Conditions) == 0 {
			newParamList = append(newParamList, param)
			continue
		}

		if _, ok := paramConditions[param.Name]; !ok {
			newParamList = append(newParamList, param)
			paramConditions[param.Name] = &conditionDefaults{}
		}
		for _, cond := range param.Conditions {
			if cond.ConditionRef == "strings-equal" {
				for _, condParam := range cond.Params {
					paramConditions[param.Name].equal = append(paramConditions[param.Name].equal, conditionDefault{key: condParam.Name, value: condParam.Value, def: param.Default})
				}
			}
		}
	}

	for i, param := range newParamList {
		if _, ok := paramConditions[param.Name]; ok {
			newParamList[i].Conditions = nil
			sortConditionalDefaults(paramConditions[param.Name].equal)
			newParamList[i].Default = paramConditions[param.Name].equal
		}
	}

	stepData.Spec.Inputs.Parameters = newParamList
}

func appendContextParameters(stepData *config.StepData) {
	contextParameterNames := stepData.GetContextParameterFilters().All
	if len(contextParameterNames) > 0 {
		contextDetailsPath := "pkg/generator/helper/piper-context-defaults.yaml"

		contextDetails := config.StepData{}
		readContextInformation(contextDetailsPath, &contextDetails)

		for _, contextParam := range contextDetails.Spec.Inputs.Parameters {
			if contains(contextParameterNames, contextParam.Name) {
				stepData.Spec.Inputs.Parameters = append(stepData.Spec.Inputs.Parameters, contextParam)
			}
		}
	}
}

func consolidateContextDefaults(stepData *config.StepData) {
	paramConditions := paramConditionDefaults{}
	for _, container := range stepData.Spec.Containers {
		containerParams := getContainerParameters(container, false)

		if container.Conditions != nil && len(container.Conditions) > 0 {
			for _, cond := range container.Conditions {
				if cond.ConditionRef == "strings-equal" {
					for _, condParam := range cond.Params {
						for paramName, val := range containerParams {
							if _, ok := paramConditions[paramName]; !ok {
								paramConditions[paramName] = &conditionDefaults{}
							}
							paramConditions[paramName].equal = append(paramConditions[paramName].equal, conditionDefault{key: condParam.Name, value: condParam.Value, def: val})
						}
					}
				}
			}
		}
	}

	stashes := []interface{}{}
	conditionalStashes := []conditionDefault{}
	for _, res := range stepData.Spec.Inputs.Resources {
		//consider only resources of type stash, others not relevant for conditions yet
		if res.Type == "stash" {
			if res.Conditions == nil || len(res.Conditions) == 0 {
				stashes = append(stashes, res.Name)
			} else {
				for _, cond := range res.Conditions {
					if cond.ConditionRef == "strings-equal" {
						for _, condParam := range cond.Params {
							conditionalStashes = append(conditionalStashes, conditionDefault{key: condParam.Name, value: condParam.Value, def: res.Name})
						}
					}
				}
			}
		}
	}

	sortConditionalDefaults(conditionalStashes)

	for _, conditionalStash := range conditionalStashes {
		stashes = append(stashes, conditionalStash)
	}

	for key, param := range stepData.Spec.Inputs.Parameters {
		if param.Name == "stashContent" {
			stepData.Spec.Inputs.Parameters[key].Default = stashes
		}

		for containerParam, paramDefault := range paramConditions {
			if param.Name == containerParam {
				sortConditionalDefaults(paramConditions[param.Name].equal)
				stepData.Spec.Inputs.Parameters[key].Default = paramDefault.equal
			}
		}
	}
}

func setDefaultAndPossisbleValues(stepData *config.StepData) {
	for k, param := range stepData.Spec.Inputs.Parameters {

		//fill default id not set
		if param.Default == nil {
			switch param.Type {
			case "bool":
				param.Default = false
			case "int":
				param.Default = 0
			}
		}

		//add possible values where known for certain types
		switch param.Type {
		case "bool":
			if param.PossibleValues == nil {
				param.PossibleValues = []interface{}{true, false}
			}
		}

		stepData.Spec.Inputs.Parameters[k] = param
	}
}

func getContainerParameters(container config.Container, sidecar bool) map[string]interface{} {
	containerParams := map[string]interface{}{}

	if len(container.Command) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarCommand", "containerCommand")] = container.Command[0]
	}
	if len(container.EnvVars) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarEnvVars", "dockerEnvVars")] = config.EnvVarsAsMap(container.EnvVars)
	}
	containerParams[ifThenElse(sidecar, "sidecarImage", "dockerImage")] = container.Image
	containerParams[ifThenElse(sidecar, "sidecarPullImage", "dockerPullImage")] = container.ImagePullPolicy != "Never"
	if len(container.Name) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarName", "containerName")] = container.Name
		containerParams["dockerName"] = container.Name
	}
	if len(container.Options) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarOptions", "dockerOptions")] = container.Options
	}
	if len(container.WorkingDir) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarWorkspace", "dockerWorkspace")] = container.WorkingDir
	}

	if sidecar {
		if len(container.ReadyCommand) > 0 {
			containerParams["sidecarReadyCommand"] = container.ReadyCommand
		}
	} else {
		if len(container.Shell) > 0 {
			containerParams["containerShell"] = container.Shell
		}
	}

	//ToDo? add dockerVolumeBind, sidecarVolumeBind -> so far not part of config.Container

	return containerParams
}

func sortStepParameters(stepData *config.StepData, considerMandatory bool) {
	if stepData.Spec.Inputs.Parameters != nil {
		parameters := stepData.Spec.Inputs.Parameters

		if considerMandatory {
			sort.SliceStable(parameters[:], func(i, j int) bool {
				if parameters[i].Mandatory == parameters[j].Mandatory {
					return strings.Compare(parameters[i].Name, parameters[j].Name) < 0
				} else if parameters[i].Mandatory {
					return true
				}
				return false
			})
		} else {
			sort.SliceStable(parameters[:], func(i, j int) bool {
				return strings.Compare(parameters[i].Name, parameters[j].Name) < 0
			})
		}
	}
}

func sortConditionalDefaults(conditionDefaults []conditionDefault) {
	sort.SliceStable(conditionDefaults[:], func(i int, j int) bool {
		keyLess := strings.Compare(conditionDefaults[i].key, conditionDefaults[j].key) < 0
		valLess := strings.Compare(conditionDefaults[i].value, conditionDefaults[j].value) < 0
		return keyLess || keyLess && valLess
	})
}
