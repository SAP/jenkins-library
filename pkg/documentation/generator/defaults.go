package generator

import (
	"sort"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
)

func appendContextParameters(stepData *config.StepData) {
	contextParameterNames := stepData.GetContextParameterFilters().All
	if len(contextParameterNames) > 0 {
		contextDetailsPath := "pkg/generator/piper-context-defaults.yaml"

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
		} else {
			for paramName, val := range containerParams {
				if _, ok := paramConditions[paramName]; !ok {
					paramConditions[paramName] = &conditionDefaults{}
				}
				paramConditions[paramName].equal = append(paramConditions[paramName].equal, conditionDefault{def: val})
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

func sortConditionalDefaults(conditionDefaults []conditionDefault) {
	sort.SliceStable(conditionDefaults[:], func(i int, j int) bool {
		keyLess := strings.Compare(conditionDefaults[i].key, conditionDefaults[j].key) < 0
		valLess := strings.Compare(conditionDefaults[i].value, conditionDefaults[j].value) < 0
		return keyLess || keyLess && valLess
	})
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
