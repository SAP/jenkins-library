package generator

import (
	"fmt"
	"slices"

	"github.com/SAP/jenkins-library/pkg/config"
)

const configRecommendation = "We recommend to define values of [step parameters](#parameters) via [.pipeline/config.yml file](../configuration.md).<br />In this case, calling the step is essentially reduced to defining the step name.<br />Calling the step can be done either in an orchestrator specific way (e.g. via a Jenkins library step) or on the command line."

const (
	headlineDescription     = "## Description\n\n"
	headlineUsage           = "## Usage\n\n"
	headlineJenkinsPipeline = "    === \"Jenkins\"\n\n"
	headlineCommandLine     = "    === \"Command Line\"\n\n"
	headlineAzure           = "    === \"Azure DevOps\"\n\n"
	headlineGHA             = "    === \"GitHub Actions\"\n\n"
	spacingTabBox           = "        "
)

// defaultBinaryName is the name of the local binary that is used for sample generation.
var defaultBinaryName string = "piper"

// defaultLibraryName is the id of the library in the Jenkins that is used for sample generation.
var defaultLibraryName string = "piper-lib-os"

// CustomLibrarySteps holds a list of libraries with it's custom steps.
var CustomLibrarySteps = []CustomLibrary{}

// CustomLibrary represents a custom library with it's custom step names, binary name and library name.
type CustomLibrary struct {
	Name        string   `yaml:"name,omitempty"`
	BinaryName  string   `yaml:"binaryName,omitempty"`
	LibraryName string   `yaml:"libraryName,omitempty"`
	Steps       []string `yaml:"steps,omitempty"`
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
	description += `!!! tip ""` + "\n\n"
	// add Jenkins-specific information
	description += headlineJenkinsPipeline
	description += fmt.Sprintf("%v```groovy\n", spacingTabBox)
	description += fmt.Sprintf("%vlibrary('%s')\n\n", spacingTabBox, libraryName)
	description += fmt.Sprintf("%v%v script: this\n", spacingTabBox, stepData.Metadata.Name)
	description += fmt.Sprintf("%v```\n\n", spacingTabBox)

	// add Azure-specific information if activated
	if includeAzure {
		description += headlineAzure
		description += fmt.Sprintf("%v```\n", spacingTabBox)
		description += fmt.Sprintf("%vsteps:\n", spacingTabBox)
		description += fmt.Sprintf("%v  - task: piper@1\n", spacingTabBox)
		description += fmt.Sprintf("%v    name: %v\n", spacingTabBox, stepData.Metadata.Name)
		description += fmt.Sprintf("%v    inputs:\n", spacingTabBox)
		description += fmt.Sprintf("%v      stepName: %v\n", spacingTabBox, stepData.Metadata.Name)
		description += fmt.Sprintf("%v      flags: --anyStepParameter\n", spacingTabBox)
		description += fmt.Sprintf("%v```\n\n", spacingTabBox)
	}

	// add GiHub Actions specific information if activated
	if includeGHA {
		description += headlineGHA
		description += fmt.Sprintf("%v```\n", spacingTabBox)
		description += fmt.Sprintf("%vsteps:\n", spacingTabBox)
		description += fmt.Sprintf("%v  - uses: SAP/project-piper-action@releaseCommitSHA\n", spacingTabBox)
		description += fmt.Sprintf("%v    name: %v\n", spacingTabBox, stepData.Metadata.Name)
		description += fmt.Sprintf("%v    with:\n", spacingTabBox)
		description += fmt.Sprintf("%v      step-name: %v\n", spacingTabBox, stepData.Metadata.Name)
		description += fmt.Sprintf("%v      flags: --anyStepParameter\n", spacingTabBox)
		description += fmt.Sprintf("%v```\n\n", spacingTabBox)
	}

	// add command line information
	description += headlineCommandLine
	description += fmt.Sprintf("%v```sh\n", spacingTabBox)
	description += fmt.Sprintf("%v%s %v\n", spacingTabBox, binaryName, stepData.Metadata.Name)
	description += fmt.Sprintf("%v```\n\n", spacingTabBox)

	description += stepOutputs(stepData)
	return description
}

func getNames(stepName string) (string, string) {
	for _, library := range CustomLibrarySteps {
		if slices.Contains(library.Steps, stepName) {
			return library.LibraryName, library.BinaryName
		}
	}
	return defaultLibraryName, defaultBinaryName
}
