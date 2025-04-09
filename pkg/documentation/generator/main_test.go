//go:build unit

package generator

import (
	"fmt"
	"io"
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
		return io.NopCloser(strings.NewReader(meta1)), nil
	default:
		return io.NopCloser(strings.NewReader("")), fmt.Errorf("Wrong Path: %v", docTemplateFilePath)
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

func TestGetBadge(t *testing.T) {
	tt := []struct {
		in       string
		expected string
	}{
		{in: "Jenkins", expected: "[![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#)"},
		{in: "jenkins", expected: "[![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#)"},
		{in: "Azure", expected: "[![Azure only](https://img.shields.io/badge/-Azure%20only-yellowgreen)](#)"},
		{in: "azure", expected: "[![Azure only](https://img.shields.io/badge/-Azure%20only-yellowgreen)](#)"},
		{in: "Github Actions", expected: "[![Github Actions only](https://img.shields.io/badge/-Github%20Actions%20only-yellowgreen)](#)"},
		{in: "github actions", expected: "[![Github Actions only](https://img.shields.io/badge/-Github%20Actions%20only-yellowgreen)](#)"},
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, getBadge(test.in))
	}
}

func TestGetStepConditionDetails(t *testing.T) {
	tt := []struct {
		name     string
		step     config.Step
		expected string
	}{
		{name: "noCondition", step: config.Step{Conditions: []config.StepCondition{}}, expected: "**active** by default - deactivate explicitly"},
		{name: "config", step: config.Step{Conditions: []config.StepCondition{{Config: map[string][]interface{}{"configKey1": {"keyVal1", "keyVal2"}}}}}, expected: "<i>config:</i><ul><li>`configKey1`: `keyVal1`</li><li>`configKey1`: `keyVal2`</li></ul>"},
		{name: "configKey", step: config.Step{Conditions: []config.StepCondition{{ConfigKey: "configKey"}}}, expected: "<i>config key:</i>&nbsp;`configKey`<br />"},
		{name: "filePattern", step: config.Step{Conditions: []config.StepCondition{{FilePattern: "testPattern"}}}, expected: "<i>file pattern:</i>&nbsp;`testPattern`<br />"},
		{name: "filePatternFromConfig", step: config.Step{Conditions: []config.StepCondition{{FilePatternFromConfig: "patternConfigKey"}}}, expected: "<i>file pattern from config:</i>&nbsp;`patternConfigKey`<br />"},
		{name: "inactive", step: config.Step{Conditions: []config.StepCondition{{Inactive: true}}}, expected: "**inactive** by default - activate explicitly"},
		{name: "npmScript", step: config.Step{Conditions: []config.StepCondition{{NpmScript: "testScript"}}}, expected: "<i>npm script:</i>&nbsp;`testScript`<br />"},
		{name: "multiple conditions", step: config.Step{Conditions: []config.StepCondition{{ConfigKey: "configKey"}, {FilePattern: "testPattern"}}}, expected: "<i>config key:</i>&nbsp;`configKey`<br /><i>file pattern:</i>&nbsp;`testPattern`<br />"},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, getStepConditionDetails(test.step))
		})
	}
}
