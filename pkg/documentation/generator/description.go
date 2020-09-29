package generator

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
)

// BinaryName is the name of the local binary that is used for sample generation.
var BinaryName string = "piper"

// LibraryName is the id of the library in the Jenkins that is used for sample generation.
var LibraryName string = "piper-lib-os"

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
	description += "### Jenkins pipelines\n\n"
	description += fmt.Sprintf("```library('%s')\n\ngroovy\n%v script: this\n```\n", LibraryName, stepData.Metadata.Name)
	description += "### Command line\n\n"
	description += fmt.Sprintf("```sh\n%s %v\n```\n\n", BinaryName, stepData.Metadata.Name)
	description += stepOutputs(stepData)
	return description
}
