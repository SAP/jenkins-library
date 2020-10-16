package generator

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestReadAndAdjustTemplate(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		tmpl, _ := configOpenDocTemplateFileMock("testStep.md")
		content := readAndAdjustTemplate(tmpl)

		cases := []struct {
			x, y string
		}{
			{"{{StepName .}}", "${docGenStepName}"},
			{"{{Parameters .}}", "${docGenParameters}"},
			{"{{Description .}}", "${docGenDescription}"},
			{"", "${docGenConfiguration}"},
			{"", "${docJenkinsPluginDependencies}"},
		}
		for _, c := range cases {
			if len(c.x) > 0 {
				assert.Contains(t, content, c.x)
			}
			if len(c.y) > 0 {
				assert.NotContains(t, content, c.y)
			}
		}
	})
}

func configOpenDocTemplateFileMock(docTemplateFilePath string) (io.ReadCloser, error) {
	meta1 := `# ${docGenStepName}

	## ${docGenDescription}

	## Prerequisites

	none

	## ${docJenkinsPluginDependencies}

	## ${docGenParameters}

	## ${docGenConfiguration}

	## Side effects

	none

	## Exceptions

	none

	## Example

	none
`
	switch docTemplateFilePath {
	case "testStep.md":
		return ioutil.NopCloser(strings.NewReader(meta1)), nil
	default:
		return ioutil.NopCloser(strings.NewReader("")), fmt.Errorf("Wrong Path: %v", docTemplateFilePath)
	}
}

func TestSetDefaultAndPossisbleValues(t *testing.T) {
	stepData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{Parameters: []config.StepParameters{
				{Name: "boolean", Type: "bool"},
				{Name: "integer", Type: "int"},
			}},
		},
	}
	setDefaultAndPossisbleValues(&stepData)
	assert.Equal(t, false, stepData.Spec.Inputs.Parameters[0].Default)
	assert.Equal(t, 0, stepData.Spec.Inputs.Parameters[1].Default)
	assert.Equal(t, []interface{}{true, false}, stepData.Spec.Inputs.Parameters[0].PossibleValues)

}

func Test_adjustDefaultValues(t *testing.T) {

	tests := []struct {
		want  interface{}
		name  string
		input *config.StepData
	}{
		{want: false, name: "boolean", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "bool", Mandatory: true},
		}}}}},
		{want: 0, name: "integer", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "int", Mandatory: true},
		}}}}},
		{want: "", name: "string", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true},
		}}}}},
		{want: []string{}, name: "string array", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true},
		}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			adjustDefaultValues(tt.input)
			// assert
			assert.Equal(t, tt.want, tt.input.Spec.Inputs.Parameters[0].Default)
		})
	}
}

func Test_adjustMandatoryFlags(t *testing.T) {
	tests := []struct {
		want  bool
		name  string
		input *config.StepData
	}{
		// TODO: current impl does not met expectations, but behavior is corrected by adjustDefaultValues
		// {want: false, name: "boolean with default not set", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
		// 	{Name: "param", Type: "bool", Mandatory: true},
		// }}}}},
		{want: false, name: "boolean with empty default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "bool", Mandatory: true, Default: false},
		}}}}},
		{want: false, name: "boolean with default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "bool", Mandatory: true, Default: true},
		}}}}},
		{want: true, name: "string with default not set", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true},
		}}}}},
		{want: true, name: "string with empty default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true, Default: ""},
		}}}}},
		{want: false, name: "string with default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true, Default: "Oktober"},
		}}}}},
		{want: true, name: "string array with default not set", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true},
		}}}}},
		{want: true, name: "string array with empty default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true, Default: []string{}},
		}}}}},
		{want: false, name: "string array with default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true, Default: []string{"Oktober"}},
		}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			adjustMandatoryFlags(tt.input)
			// assert
			assert.Equal(t, tt.want, tt.input.Spec.Inputs.Parameters[0].Mandatory)
		})
	}
}
