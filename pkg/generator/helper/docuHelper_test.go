package helper

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
			{"{{docGenStepName .}}", "${docGenStepName}"},
			{"{{docGenConfiguration .}}", "${docGenConfiguration}"},
			{"{{docGenParameters .}}", "${docGenParameters}"},
			{"{{docGenDescription .}}", "${docGenDescription}"},
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

func TestCreateParameterOverview(t *testing.T) {
	stepData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Resources: []config.StepResources{
					{Name: "testStash", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{Name: "param1"},
					{Name: "stashContent", Default: "testStash"},
				},
			},
		},
	}

	expected := `| Name | Mandatory | Additional information |
| ---- | --------- | ---------------------- |
| [param1](#param1) | no |  |
| [stashContent](#stashContent) | no | [![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#) |

`

	assert.Equal(t, expected, createParameterOverview(&stepData))
}

func TestParameterFurtherInfo(t *testing.T) {
	tt := []struct {
		paramName     string
		contextParams []string
		stepData      *config.StepData
		contains      string
		notContains   []string
	}{
		{paramName: "verbose", contextParams: []string{}, stepData: nil, contains: "activates debug output"},
		{paramName: "script", contextParams: []string{}, stepData: nil, contains: "reference to Jenkins main pipeline script"},
		{paramName: "contextTest", contextParams: []string{"contextTest"}, stepData: &config.StepData{}, contains: "Jenkins only", notContains: []string{"pipeline script", "id of credentials"}},
		{paramName: "noop", contextParams: []string{}, stepData: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{}}}}, contains: ""},
		{
			paramName:     "testCredentialId",
			contextParams: []string{"testCredentialId"},
			stepData: &config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Secrets: []config.StepSecrets{{Name: "testCredentialId", Type: "jenkins"}},
					},
				},
			},
			contains: "id of credentials",
		},
		{
			paramName: "testSecret",
			stepData: &config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testSecret", Secret: true, ResourceRef: []config.ResourceReference{{Name: "mytestSecret", Type: "secret"}}},
						},
					},
				},
			},
			contains: "credentials ([`mytestSecret`](#mytestSecret))",
		},
		{
			paramName: "testSecret",
			stepData: &config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testSecret"},
						},
					},
				},
			},
			contains: "",
		},
	}

	for _, test := range tt {
		res := parameterFurtherInfo(test.paramName, test.contextParams, test.stepData)
		if len(test.contains) == 0 {
			assert.Equal(t, test.contains, res)
		} else {
			assert.Contains(t, res, test.contains)
		}
		for _, notThere := range test.notContains {
			assert.NotContains(t, res, notThere)
		}
	}
}

func TestCreateParameterDetails(t *testing.T) {
	stepData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:            "param1",
						Aliases:         []config.Alias{{Name: "param1Alias"}, {Name: "paramAliasDeprecated", Deprecated: true}},
						Mandatory:       true,
						Default:         "param1Default",
						LongDescription: "long description",
						PossibleValues:  []interface{}{"val1", "val2"},
						Scope:           []string{"STEPS"},
						Secret:          true,
						Type:            "string",
					},
				},
			},
		},
	}

	res := createParameterDetails(&stepData)

	assert.Contains(t, res, "#### param1")
	assert.Contains(t, res, "long description")
	assert.Contains(t, res, "`param1Alias`")
	assert.Contains(t, res, "`paramAliasDeprecated` (**deprecated**)")
	assert.Contains(t, res, "string")
	assert.Contains(t, res, "param1Default")
	assert.Contains(t, res, "val1")
	assert.Contains(t, res, "val2")
	assert.Contains(t, res, "no")
	assert.Contains(t, res, "**yes**")
	assert.Contains(t, res, "steps")
}

func TestConsolidateConditionalParameters(t *testing.T) {
	stepData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{Name: "dep1", Default: "val1"},
					{Name: "test1", Default: "def1", Conditions: []config.Condition{
						{ConditionRef: "strings-equal", Params: []config.Param{{Name: "dep1", Value: "val1"}, {Name: "dep2", Value: "val1"}}},
					}},
					{Name: "test1", Default: "def2", Conditions: []config.Condition{
						{ConditionRef: "strings-equal", Params: []config.Param{{Name: "dep1", Value: "val2"}, {Name: "dep2", Value: "val2"}}},
					}},
				},
			},
		},
	}

	consolidateConditionalParameters(&stepData)

	expected := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{Name: "dep1", Default: "val1"},
					{Name: "test1", Default: []conditionDefault{
						{key: "dep1", value: "val1", def: "def1"},
						{key: "dep1", value: "val2", def: "def2"},
						{key: "dep2", value: "val1", def: "def1"},
						{key: "dep2", value: "val2", def: "def2"},
					}},
				},
			},
		},
	}

	assert.Equal(t, expected, stepData)

}

func TestConsolidateContextParameters(t *testing.T) {
	stepData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{Name: "stashContent"},
					{Name: "dockerImage"},
					{Name: "containerName"},
					{Name: "dockerName"},
				},
				Resources: []config.StepResources{
					{Name: "stashAlways", Type: "stash"},
					{Name: "stash1", Type: "stash", Conditions: []config.Condition{
						{ConditionRef: "strings-equal", Params: []config.Param{{Name: "dep1", Value: "val1"}, {Name: "dep2", Value: "val1"}}},
					}},
					{Name: "stash2", Type: "stash", Conditions: []config.Condition{
						{ConditionRef: "strings-equal", Params: []config.Param{{Name: "dep1", Value: "val2"}, {Name: "dep2", Value: "val2"}}},
					}},
				},
			},
			Containers: []config.Container{
				{Name: "IMAGE1", Image: "image1", Conditions: []config.Condition{
					{ConditionRef: "strings-equal", Params: []config.Param{{Name: "dep1", Value: "val1"}, {Name: "dep2", Value: "val1"}}},
				}},
				{Name: "IMAGE2", Image: "image2", Conditions: []config.Condition{
					{ConditionRef: "strings-equal", Params: []config.Param{{Name: "dep1", Value: "val2"}, {Name: "dep2", Value: "val2"}}},
				}},
			},
		},
	}

	consolidateContextDefaults(&stepData)

	expected := []config.StepParameters{
		{Name: "stashContent", Default: []interface{}{
			"stashAlways",
			conditionDefault{key: "dep1", value: "val1", def: "stash1"},
			conditionDefault{key: "dep1", value: "val2", def: "stash2"},
			conditionDefault{key: "dep2", value: "val1", def: "stash1"},
			conditionDefault{key: "dep2", value: "val2", def: "stash2"},
		}},
		{Name: "dockerImage", Default: []conditionDefault{
			{key: "dep1", value: "val1", def: "image1"},
			{key: "dep1", value: "val2", def: "image2"},
			{key: "dep2", value: "val1", def: "image1"},
			{key: "dep2", value: "val2", def: "image2"},
		}},
		{Name: "containerName", Default: []conditionDefault{
			{key: "dep1", value: "val1", def: "IMAGE1"},
			{key: "dep1", value: "val2", def: "IMAGE2"},
			{key: "dep2", value: "val1", def: "IMAGE1"},
			{key: "dep2", value: "val2", def: "IMAGE2"},
		}},
		{Name: "dockerName", Default: []conditionDefault{
			{key: "dep1", value: "val1", def: "IMAGE1"},
			{key: "dep1", value: "val2", def: "IMAGE2"},
			{key: "dep2", value: "val1", def: "IMAGE1"},
			{key: "dep2", value: "val2", def: "IMAGE2"},
		}},
	}

	assert.Equal(t, expected, stepData.Spec.Inputs.Parameters)

}

var expectedResultDocument string = "# testStep\n\n\t## Description\n\nLong Test description\n\n\t\n\t" +
	"## Prerequisites\n\t\n\tnone\n\n\t\n\t\n\t" +
	"## Parameters\n\n| name | mandatory | default | possible values |\n" +
	"| ---- | --------- | ------- | --------------- |\n" +
	"| `param0` | No | `val0` |  |\n" +
	"| `param1` | No |  |  |\n" +
	"| `param2` | Yes |  |  |\n" +
	"| `script` | Yes |  |  |\n" +
	"| `verbose` | No | `false` | `true`, `false` |\n\n" +
	" * `param0`: param0 description\n" +
	" * `param1`: param1 description\n" +
	" * `param2`: param1 description\n" +
	" * `script`: The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` for retrieving, e.g. configuration parameters.\n" +
	" * `verbose`: verbose output\n\n\t\n\t" +
	"## Step Configuration\n\nWe recommend to define values of step parameters via [config.yml file](../configuration.md).\n\n" +
	"In following sections of the config.yml the configuration is possible:\n\n" +
	"| parameter | general | step/stage |\n" +
	"| --------- | ------- | ---------- |\n" +
	"| `param0` | X |  |\n" +
	"| `param1` |  |  |\n" +
	"| `param2` |  |  |\n" +
	"| `verbose` | X |  |\n\n\t\n\t" +
	"## Side effects\n\t\n\tnone\n\t\n\t" +
	"## Exceptions\n\t\n\tnone\n\t\n\t" +
	"## Example\n\n\tnone\n"

func configMetaDataMock(name string) (io.ReadCloser, error) {
	meta1 := `metadata:
  name: testStep
  description: Test description
  longDescription: |
    Long Test description
spec:
  inputs:
    params:
      - name: param0
        type: string
        description: param0 description
        default: val0
        scope:
        - GENERAL
        - PARAMETERS
        mandatory: true
      - name: param1
        type: string
        description: param1 description
        scope:
        - PARAMETERS
      - name: param2
        type: string
        description: param1 description
        scope:
        - PARAMETERS
        mandatory: true
`
	var r string
	switch name {
	case "test.yaml":
		r = meta1
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
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

var resultDocumentContent string

func docFileWriterMock(docTemplateFilePath string, data []byte, perm os.FileMode) error {

	resultDocumentContent = string(data)
	switch docTemplateFilePath {
	case "testStep.md":
		return nil
	default:
		return fmt.Errorf("Wrong Path: %v", docTemplateFilePath)
	}
}

func TestSortStepParameters(t *testing.T) {
	stepData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{Name: "ab1", Mandatory: false},
					{Name: "ab3", Mandatory: true},
					{Name: "ab5", Mandatory: false},
					{Name: "ab6", Mandatory: true},
					{Name: "ab4", Mandatory: false},
					{Name: "ab2", Mandatory: true},
				},
			},
		},
	}

	t.Run("ignore mandatory", func(t *testing.T) {
		sortStepParameters(&stepData, false)

		assert.Equal(t, "ab1", stepData.Spec.Inputs.Parameters[0].Name)
		assert.Equal(t, "ab2", stepData.Spec.Inputs.Parameters[1].Name)
		assert.Equal(t, "ab3", stepData.Spec.Inputs.Parameters[2].Name)
		assert.Equal(t, "ab4", stepData.Spec.Inputs.Parameters[3].Name)
		assert.Equal(t, "ab5", stepData.Spec.Inputs.Parameters[4].Name)
		assert.Equal(t, "ab6", stepData.Spec.Inputs.Parameters[5].Name)
	})

	t.Run("consider mandatory", func(t *testing.T) {
		sortStepParameters(&stepData, true)

		assert.Equal(t, "ab2", stepData.Spec.Inputs.Parameters[0].Name)
		assert.Equal(t, "ab3", stepData.Spec.Inputs.Parameters[1].Name)
		assert.Equal(t, "ab6", stepData.Spec.Inputs.Parameters[2].Name)
		assert.Equal(t, "ab1", stepData.Spec.Inputs.Parameters[3].Name)
		assert.Equal(t, "ab4", stepData.Spec.Inputs.Parameters[4].Name)
		assert.Equal(t, "ab5", stepData.Spec.Inputs.Parameters[5].Name)
	})
}
