package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
)

// Replaces the Parameters placeholder with the content from the yaml
func createParametersSection(stepData *config.StepData) string {

	var parameters = "## Parameters\n\n"

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
		details += fmt.Sprintf("| Mandatory | %v |\n", ifThenElse(param.Mandatory, "**yes**", "no"))
		details += fmt.Sprintf("| Default | %v |\n", formatDefault(param, stepParameterNames))
		if param.PossibleValues != nil {
			details += fmt.Sprintf("| Possible values | %v |\n", possibleValueList(param.PossibleValues))
		}
		details += fmt.Sprintf("| Secret | %v |\n", ifThenElse(param.Secret, "**yes**", "no"))
		details += fmt.Sprintf("| Configuration scope | %v |\n", scopeDetails(param.Scope))
		details += fmt.Sprintf("| Resource references | %v |\n", resourceReferenceDetails(param.ResourceRef))

		details += "\n\n"
	}

	for _, secret := range stepData.Spec.Inputs.Secrets {
		details += fmt.Sprintf("#### %v\n\n", secret.Name)

		if !contains(stepParameterNames, secret.Name) {
			details += "**Jenkins-specific:** Used for proper environment setup. See *[using credentials](https://www.jenkins.io/doc/book/using/using-credentials/)* for details.\n\n"
		}

		details += secret.Description + "\n\n"

		details += "[back to overview](#parameters)\n\n"

		details += "| Scope | Details |\n"
		details += "| ---- | --------- |\n"
		details += fmt.Sprintf("| Aliases | %v |\n", aliasList(secret.Aliases))
		details += fmt.Sprintf("| Type | `%v` |\n", "string")
		details += fmt.Sprintf("| Configuration scope | %v |\n", scopeDetails([]string{"PARAMETERS", "GENERAL", "STEPS", "STAGES"}))

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
			resourceDetails += fmt.Sprintf("&nbsp;&nbsp;id: [`%v`](#%v)<br />", resource.Name, strings.ToLower(resource.Name))
			resourceDetails += fmt.Sprintf("&nbsp;&nbsp;reference to: `%v`<br />", resource.Param)
			continue
		}
	}

	return resourceDetails
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
