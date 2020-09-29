package generator

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
)

const configRecommendation = "We recommend to define values of [step parameters](#parameters) via [config.yml file](../configuration.md). In this case, calling the step is reduced to one simple line.<br />Calling the step can be done either via the Jenkins library step or on the [command line](../cli/index.md)."

const (
	headlineDescription     = "## Description\n\n"
	headlineUsage           = "## Usage\n\n"
	headlineJenkinsPipeline = "### Jenkins Pipeline\n\n"
	headlineCommandLine     = "### Command Line\n\n"
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

	description += headlineDescription + stepData.Metadata.LongDescription + "\n\n"

	description += headlineUsage
	description += configRecommendation + "\n\n"
	description += headlineJenkinsPipeline
	description += fmt.Sprintf("```groovy\nlibrary('%s')\n\n%v script: this\n```\n\n", LibraryName, stepData.Metadata.Name)
	description += headlineCommandLine
	description += fmt.Sprintf("```sh\n%s %v\n```\n\n", BinaryName, stepData.Metadata.Name)
	description += stepOutputs(stepData)
	return description
}
