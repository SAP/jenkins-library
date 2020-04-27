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

var expectedResultDocument string = "# testStep\n\n\t## Description\n\nLong Test description\n\n\t\n\t## Prerequisites\n\t\n\tnone\n\n\t\n\t\n\t## Parameters\n\n| name | mandatory | default | possible values |\n| ---- | --------- | ------- | --------------- |\n| `param0` | No | `val0` |  |\n| `param1` | No |  |  |\n| `param2` | Yes |  |  |\n\n * `param0`: param0 description\n * `param1`: param1 description\n * `param2`: param1 description\n\n\t\n\t## Step Configuration\n\nWe recommend to define values of step parameters via [config.yml file](../configuration.md).\n\nIn following sections of the config.yml the configuration is possible:\n\n| parameter | general | step/stage |\n| --------- | ------- | ---------- |\n| `param0` | X |  |\n| `param1` |  |  |\n| `param2` |  |  |\n\n\t\n\t## Side effects\n\t\n\tnone\n\t\n\t## Exceptions\n\t\n\tnone\n\t\n\t## Example\n\n\tnone\n"

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

func TestAddDefaultContainerContent(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)
		addDefaultContainerContent(&stepData, m)

		cases := []struct {
			x, want string
		}{
			{"containerCommand", "`command`"},
			{"containerName", "`container0`, `container1``container2a`<br>`container2b`<br>"},
			{"containerShell", "`shell`"},
			{"dockerEnvVars", "`[envar.name0=envar.value0]`, `[envar.name1=envar.value1]`param_name2a=`param_value2a`: `[envar.name2a=envar.value2a]`<br>param.name2b=`param.value2b`: `[envar.name2b=envar.value2b]`<br>"},
			{"dockerImage", "`image`, `image`param_name2a=`param_value2a`: `image`<br>param.name2b=`param.value2b`: `image`<br>"},
			{"dockerName", "`container0`, `container1``container2a`<br>`container2b`<br>"},
			{"dockerPullImage", "true"},
			{"dockerOptions", "option.name2b option.value2b"},
			{"dockerWorkspace", "`workingdir`, `workingdir`param_name2a=`param_value2a`: `workingdir`<br>param.name2b=`param.value2b`: `workingdir`<br>"},
		}
		//assert.Equal(t, len(cases), len(m))
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v: %v != %v", c.x, m[c.x], c.want))
		}
	})
}
func TestAddDefaultSidecarContent(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)
		addDefaultSidecarContent(&stepData, m)

		cases := []struct {
			x, want string
		}{
			{"sidecarCommand", "command"},
			{"sidecarEnvVars", "envar.name3=envar.value3"},
			{"sidecarImage", "`image`"},
			{"sidecarName", "`sidecar0`"},
			{"sidecarPullImage", "true"},
			{"sidecarReadyCommand", "readycommand"},
			{"sidecarOptions", "option.name3b option.value3b"},
			{"sidecarWorkspace", "workingdir"},
		}
		assert.Equal(t, len(cases), len(m))
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.Equal(t, c.want, m[c.x], fmt.Sprintf("%v:%v", c.x, m[c.x]))
		}
	})
}

func TestAddStashContent(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)
		addStashContent(&stepData, m)

		cases := []struct {
			x, want string
		}{
			{"stashContent", "resource0, resource1, resource2"},
		}
		assert.Equal(t, len(cases), len(m))
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v:%v", c.x, m[c.x]))
		}
	})
}

func TestGetDocuContextDefaults(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		m := getDocuContextDefaults(&stepData)

		cases := []struct {
			x, want string
		}{
			{"stashContent", "resource0, resource1, resource2"},
			{"sidecarCommand", "command"},
			{"sidecarEnvVars", "envar.name3=envar.value3"},
			{"sidecarImage", "image"},
			{"sidecarName", "sidecar0"},
			{"sidecarPullImage", "true"},
			{"sidecarReadyCommand", "readycommand"},
			{"sidecarOptions", "option.name3b option.value3b"},
			{"sidecarWorkspace", "workingdir"},
			{"containerCommand", "command"},
			{"containerName", "`container0`, `container1``container2a`<br>`container2b`<br>"},
			{"containerShell", "shell"},
			{"dockerEnvVars", "`[envar.name0=envar.value0]`, `[envar.name1=envar.value1]`param_name2a=`param_value2a`: `[envar.name2a=envar.value2a]`<br>param.name2b=`param.value2b`: `[envar.name2b=envar.value2b]`<br>"},
			{"dockerImage", "`image`, `image`param_name2a=`param_value2a`: `image`<br>param.name2b=`param.value2b`: `image`"},
			{"dockerName", "`container0`, `container1``container2a`<br>`container2b`<br>"},
			{"dockerPullImage", "true"},
			{"dockerOptions", "option.name2b option.value2b"},
			{"dockerWorkspace", "`workingdir`, `workingdir`param_name2a=`param_value2a`: `workingdir`<br>param.name2b=`param.value2b`: `workingdir`<br>"},
		}
		assert.Equal(t, len(cases), len(m))
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v: %v != %v", c.x, m[c.x], c.want))
		}
	})
}

func TestAddNewParameterWithCondition(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)

		cases := []struct {
			x, want string
			i       int
		}{
			{"param0", "name0a=`val0a`: `default0` name0b=`val0b`: `default0`", 0},
			{"param1", "name1a=`val1a`: `default1`", 1},
		}
		for _, c := range cases {

			addNewParameterWithCondition(stepData.Spec.Inputs.Parameters[c.i], m)
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v", m[c.x]))
		}
	})
}

func TestAddExistingParameterWithCondition(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)
		addNewParameterWithCondition(stepData.Spec.Inputs.Parameters[1], m)

		cases := []struct {
			x, want string
		}{
			{"param1", "name1a=`val1a`: `default1` <br>name1b=`val1b`: `default1` "},
		}
		for _, c := range cases {

			addExistingParameterWithCondition(stepData.Spec.Inputs.Parameters[2], m)
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v", m[c.x]))
		}
	})
}

func TestRenderPossibleValues(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		// init
		var in []interface{}
		// test
		out := possibleValuesToString(in)
		// assert
		assert.Empty(t, out)
	})
	t.Run("one", func(t *testing.T) {
		// init
		var in []interface{}
		in = append(in, "fu")
		// test
		out := possibleValuesToString(in)
		// assert
		assert.Equal(t, "`fu`", out)
	})
	t.Run("many", func(t *testing.T) {
		// init
		var in []interface{}
		in = append(in, "fu", "fara")
		// test
		out := possibleValuesToString(in)
		// assert
		assert.Equal(t, "`fu`, `fara`", out)
	})
	t.Run("boolean", func(t *testing.T) {
		// init
		var in []interface{}
		in = append(in, false, true)
		// test
		out := possibleValuesToString(in)
		// assert
		assert.Equal(t, "`false`, `true`", out)
	})
}
