package generator

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
)

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
