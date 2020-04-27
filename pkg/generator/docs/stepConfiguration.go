package docs

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	SliceUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

//BuildStepConfiguration creates a string with the content from the step configuration data.
func BuildStepConfiguration(stepData config.StepData) (content string) {
	content = "Step Configuration\n\n"
	content += "We recommend to define values of step parameters via [config.yml file](../configuration.md).\n\n"
	content += "In following sections of the config.yml the configuration is possible:\n\n"
	content += "| parameter | general | step/stage |\n"
	content += "| --------- | ------- | ---------- |\n"

	for _, parameter := range stepData.Spec.Inputs.Parameters {
		if len(parameter.Scope) > 0 {
			content += fmt.Sprintf("| `%v` | %v | %v |\n",
				parameter.Name,
				ifThenElse(SliceUtils.ContainsString(parameter.Scope, "GENERAL"), "X", ""),
				ifThenElse(SliceUtils.ContainsString(parameter.Scope, "STEPS"), "X", ""))
		}
	}
	return
}
