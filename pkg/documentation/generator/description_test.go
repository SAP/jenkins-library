//go:build unit
// +build unit

package generator

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateStepName(t *testing.T) {
	tests := []struct {
		name  string
		input *config.StepData
		want  string
	}{
		{
			name: "simple step name section",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "teststep", Description: "TestDescription"},
			},
			want: "# teststep\n\nTestDescription\n",
		},
		{
			name: "step name with jenkins orchestrator uses yellowgreen badge",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "teststep", Description: "TestDescription", Orchestrators: []string{"jenkins"}},
			},
			want: "# teststep [![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#)\n\nTestDescription\n",
		},
		{
			name: "step name with gha orchestrator uses blue badge",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "teststep", Description: "TestDescription", Orchestrators: []string{"gha"}},
			},
			want: "# teststep [![GitHub Actions only](https://img.shields.io/badge/-GitHub%20Actions%20only-blue)](#)\n\nTestDescription\n",
		},
		{
			name: "step name with azure orchestrator uses light blue badge",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "teststep", Description: "TestDescription", Orchestrators: []string{"azure"}},
			},
			want: "# teststep [![Azure DevOps only](https://img.shields.io/badge/-Azure%20DevOps%20only-9cf)](#)\n\nTestDescription\n",
		},
		{
			name: "step name with multiple orchestrators renders no badge",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "teststep", Description: "TestDescription", Orchestrators: []string{"jenkins", "gha"}},
			},
			want: "# teststep\n\nTestDescription\n",
		},
	}
	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.want, createStepName(testcase.input))
		})
	}
}

func TestCreateDescriptionSection(t *testing.T) {
	CustomLibrarySteps = []CustomLibrary{{
		Name:        "TestLibrary",
		BinaryName:  "myBinary",
		LibraryName: "myLibrary",
		Steps:       []string{"myCustomStep"},
	}}

	tests := []struct {
		name  string
		input *config.StepData
		want  string
	}{
		{
			name: "simple step description section",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "teststep", LongDescription: "TestDescription"},
			},
			want: headlineDescription + "TestDescription" + "\n\n" +
				headlineUsage + configRecommendation + "\n\n" +
				"!!! tip \"\"" + "\n\n" +
				headlineJenkinsPipeline + "        ```groovy\n        library('piper-lib-os')\n\n        teststep script: this\n        ```" + "\n\n" +
				headlineCommandLine + "        ```sh\n        piper teststep\n        ```" + "\n\n",
		},
		{
			name: "custom step description section",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "myCustomStep", LongDescription: "TestDescription"},
			},
			want: headlineDescription + "TestDescription" + "\n\n" +
				headlineUsage + configRecommendation + "\n\n" +
				"!!! tip \"\"" + "\n\n" +
				headlineJenkinsPipeline + "        ```groovy\n        library('myLibrary')\n\n        myCustomStep script: this\n        ```" + "\n\n" +
				headlineCommandLine + "        ```sh\n        myBinary myCustomStep\n        ```" + "\n\n",
		},
	}
	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.want, createDescriptionSection(testcase.input))
		})
	}
}
