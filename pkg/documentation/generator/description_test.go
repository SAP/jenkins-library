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
				headlineJenkinsPipeline + "```groovy\nlibrary('piper-lib-os')\n\nteststep script: this\n```" + "\n\n" +
				headlineCommandLine + "```sh\npiper teststep\n```" + "\n\n",
		},
		{
			name: "custom step description section",
			input: &config.StepData{
				Metadata: config.StepMetadata{Name: "myCustomStep", LongDescription: "TestDescription"},
			},
			want: headlineDescription + "TestDescription" + "\n\n" +
				headlineUsage + configRecommendation + "\n\n" +
				headlineJenkinsPipeline + "```groovy\nlibrary('myLibrary')\n\nmyCustomStep script: this\n```" + "\n\n" +
				headlineCommandLine + "```sh\nmyBinary myCustomStep\n```" + "\n\n",
		},
	}
	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.want, createDescriptionSection(testcase.input))
		})
	}
}
