package generator

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
)

// Replaces the StepName placeholder with the content from the yaml
func createStepName(stepData *config.StepData) string {
	return stepData.Metadata.Name + "\n\n" + stepData.Metadata.Description + "\n"
}

// Replaces the Description placeholder with content from the yaml
func createDescriptionSection(stepData *config.StepData) string {
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
