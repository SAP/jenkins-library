package generator

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
)

func stepOutputs(stepData *config.StepData) string {
	if len(stepData.Spec.Outputs.Resources) == 0 {
		return ""
	}

	var stepOutput strings.Builder
	stepOutput.WriteString("\n## Outputs\n\n")
	stepOutput.WriteString("| Output type | Details |\n")
	stepOutput.WriteString("| ----------- | ------- |\n")

	for _, res := range stepData.Spec.Outputs.Resources {
		//handle commonPipelineEnvironment output
		if res.Type == "piperEnvironment" {
			stepOutput.WriteString(fmt.Sprintf("| %v | <ul>", res.Name))
			for _, param := range res.Parameters {
				stepOutput.WriteString(fmt.Sprintf("<li>%v</li>", param["name"]))
			}
			stepOutput.WriteString("</ul> |\n")
		}

		//handle Influx output
		if res.Type == "influx" {
			stepOutput.WriteString(fmt.Sprintf("| %v | ", res.Name))
			for _, param := range res.Parameters {
				stepOutput.WriteString(fmt.Sprintf("measurement `%v`<br /><ul>", param["name"]))
				fields, _ := param["fields"].([]any)
				for _, field := range fields {
					fieldMap, _ := field.(map[string]any)
					stepOutput.WriteString(fmt.Sprintf("<li>%v</li>", fieldMap["name"]))
				}
			}
			stepOutput.WriteString("</ul> |\n")
		}

	}
	return stepOutput.String()
}
