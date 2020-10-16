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
