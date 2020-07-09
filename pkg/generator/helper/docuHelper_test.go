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

func TestConsolidateConditionalParameters(t *testing.T) {
	stepData = config.StepData{
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
	stepData = config.StepData{
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

var stepData config.StepData = config.StepData{
	Spec: config.StepSpec{
		Inputs: config.StepInputs{
			Parameters: []config.StepParameters{
				{Name: "param0", Scope: []string{"GENERAL"}, Type: "string", Default: "default0",
					Conditions: []config.Condition{
						{Params: []config.Param{
							{"name0a", "val0a"},
							{"name0b", "val0b"},
						},
						}},
				},
				{Name: "param1", Scope: []string{"GENERAL"}, Type: "string", Default: "default1",
					Conditions: []config.Condition{
						{Params: []config.Param{
							{"name1a", "val1a"},
						},
						}},
				},
				{Name: "param1", Scope: []string{"GENERAL"}, Type: "string", Default: "default1",
					Conditions: []config.Condition{
						{Params: []config.Param{
							{"name1b", "val1b"},
						},
						}},
				},
			},
			Resources: []config.StepResources{
				{Name: "resource0", Type: "stash", Description: "val0"},
				{Name: "resource1", Type: "stash", Description: "val1"},
				{Name: "resource2", Type: "stash", Description: "val2"},
			},
		},
		Containers: []config.Container{
			{Name: "container0", Image: "image", WorkingDir: "workingdir", Shell: "shell",
				EnvVars: []config.EnvVar{
					{"envar.name0", "envar.value0"},
				},
			},
			{Name: "container1", Image: "image", WorkingDir: "workingdir",
				EnvVars: []config.EnvVar{
					{"envar.name1", "envar.value1"},
				},
			},
			{Name: "container2a", Command: []string{"command"}, ImagePullPolicy: "pullpolicy", Image: "image", WorkingDir: "workingdir",
				EnvVars: []config.EnvVar{
					{"envar.name2a", "envar.value2a"}},
				Conditions: []config.Condition{
					{Params: []config.Param{
						{"param_name2a", "param_value2a"},
					}},
				},
			},
			{Name: "container2b", Image: "image", WorkingDir: "workingdir",
				EnvVars: []config.EnvVar{
					{"envar.name2b", "envar.value2b"},
				},
				Conditions: []config.Condition{
					{Params: []config.Param{
						{"param.name2b", "param.value2b"},
					}},
				},
				//VolumeMounts: []config.VolumeMount{
				//	{"mp.2b", "mn.2b"},
				//},
				Options: []config.Option{
					{"option.name2b", "option.value2b"},
				},
			},
		},
		Sidecars: []config.Container{
			{Name: "sidecar0", Command: []string{"command"}, ImagePullPolicy: "pullpolicy", Image: "image", WorkingDir: "workingdir", ReadyCommand: "readycommand",
				EnvVars: []config.EnvVar{
					{"envar.name3", "envar.value3"}},
				Conditions: []config.Condition{
					{Params: []config.Param{
						{"param.name0", "param.value0"},
					}},
				},
				//VolumeMounts: []config.VolumeMount{
				//	{"mp.3b", "mn.3b"},
				//},
				Options: []config.Option{
					{"option.name3b", "option.value3b"},
				},
			},
		},
	},
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

func TestGenerateStepDocumentationSuccess(t *testing.T) {
	var stepData config.StepData
	contentMetaData, _ := configMetaDataMock("test.yaml")
	stepData.ReadPipelineStepData(contentMetaData)

	generateStepDocumentation(stepData, DocuHelperData{true, "", configOpenDocTemplateFileMock, docFileWriterMock})

	t.Run("Docu Generation Success", func(t *testing.T) {
		assert.Equal(t, expectedResultDocument, resultDocumentContent)
	})
}

func TestGenerateStepDocumentationError(t *testing.T) {
	var stepData config.StepData
	contentMetaData, _ := configMetaDataMock("test.yaml")
	stepData.ReadPipelineStepData(contentMetaData)

	err := generateStepDocumentation(stepData, DocuHelperData{true, "Dummy", configOpenDocTemplateFileMock, docFileWriterMock})

	t.Run("Docu Generation Success", func(t *testing.T) {
		assert.Error(t, err, fmt.Sprintf("Error occured: %v\n", err))
	})
}

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
