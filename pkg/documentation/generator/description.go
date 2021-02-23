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

// defaultBinaryName is the name of the local binary that is used for sample generation.
var defaultBinaryName string = "piper"

// defaultLibraryName is the id of the library in the Jenkins that is used for sample generation.
var defaultLibraryName string = "piper-lib-os"

// CustomLibrarySteps holds a list of libraries with it's custom steps.
var CustomLibrarySteps = []CustomLibrary{}

// CustomLibrary represents a custom library with it's custom step names, binary name and library name.
type CustomLibrary struct {
	Name        string   `yaml: "name,omitempty"`
	BinaryName  string   `yaml: "binaryName,omitempty"`
	LibraryName string   `yaml: "libraryName,omitempty"`
	Steps       []string `yaml: "steps,omitempty"`
}

// Replaces the StepName placeholder with the content from the yaml
func createStepName(stepData *config.StepData) string {
	return "# " + stepData.Metadata.Name + "\n\n" + stepData.Metadata.Description + "\n"
}

// Replaces the Description placeholder with content from the yaml
func createDescriptionSection(stepData *config.StepData) string {
	libraryName, binaryName := getNames(stepData.Metadata.Name)

	description := ""
	description += headlineDescription + stepData.Metadata.LongDescription + "\n\n"
	description += headlineUsage
	description += configRecommendation + "\n\n"
	description += headlineJenkinsPipeline
	description += fmt.Sprintf("```groovy\nlibrary('%s')\n\n%v script: this\n```\n\n", libraryName, stepData.Metadata.Name)
	description += headlineCommandLine
	description += fmt.Sprintf("```sh\n%s %v\n```\n\n", binaryName, stepData.Metadata.Name)
	description += stepOutputs(stepData)
	return description
}

func getNames(stepName string) (string, string) {
	for _, library := range CustomLibrarySteps {
		for _, customStepName := range library.Steps {
			if stepName == customStepName {
				return library.LibraryName, library.BinaryName
			}
		}
	}
	return defaultLibraryName, defaultBinaryName
}
