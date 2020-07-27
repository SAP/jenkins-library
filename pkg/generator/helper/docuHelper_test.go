package helper

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

func TestStepOutputs(t *testing.T) {
	t.Run("no resources", func(t *testing.T) {
		stepData := config.StepData{Spec: config.StepSpec{Outputs: config.StepOutputs{Resources: []config.StepResources{}}}}
		result := stepOutputs(&stepData)
		assert.Equal(t, "", result)
	})

	t.Run("with resources", func(t *testing.T) {
		stepData := config.StepData{Spec: config.StepSpec{Outputs: config.StepOutputs{Resources: []config.StepResources{
			{Name: "commonPipelineEnvironment", Type: "piperEnvironment", Parameters: []map[string]interface{}{{"name": "param1"}, {"name": "param2"}}},
			{
				Name: "influxName",
				Type: "influx",
				Parameters: []map[string]interface{}{
					{"name": "influx1", "fields": []interface{}{
						map[string]interface{}{"name": "1_1"},
						map[string]interface{}{"name": "1_2"},
					}},
					{"name": "influx2", "fields": []interface{}{
						map[string]interface{}{"name": "2_1"},
						map[string]interface{}{"name": "2_2"},
					}},
				},
			},
		}}}}
		result := stepOutputs(&stepData)
		assert.Contains(t, result, "## Outputs")
		assert.Contains(t, result, "| influxName |")
		assert.Contains(t, result, "measurement `influx1`<br /><ul>")
		assert.Contains(t, result, "measurement `influx2`<br /><ul>")
		assert.Contains(t, result, "<li>1_1</li>")
		assert.Contains(t, result, "<li>1_2</li>")
		assert.Contains(t, result, "<li>2_1</li>")
		assert.Contains(t, result, "<li>2_2</li>")
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
| [stashContent](#stashcontent) | no | [![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)](#) |

`
	stepParameterNames = []string{"param1"}
	assert.Equal(t, expected, createParameterOverview(&stepData))
	stepParameterNames = []string{}
}

func TestParameterFurtherInfo(t *testing.T) {
	tt := []struct {
		paramName   string
		stepParams  []string
		stepData    *config.StepData
		contains    string
		notContains []string
	}{
		{paramName: "verbose", stepParams: []string{}, stepData: nil, contains: "activates debug output"},
		{paramName: "script", stepParams: []string{}, stepData: nil, contains: "reference to Jenkins main pipeline script"},
		{paramName: "contextTest", stepParams: []string{}, stepData: &config.StepData{}, contains: "Jenkins only", notContains: []string{"pipeline script", "id of credentials"}},
		{paramName: "noop", stepParams: []string{"noop"}, stepData: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{}}}}, contains: ""},
		{
			paramName: "testCredentialId",
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
			paramName:  "testSecret1",
			stepParams: []string{"testSecret1"},
			stepData: &config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testSecret1", Secret: true, ResourceRef: []config.ResourceReference{{Name: "mytestSecret", Type: "secret"}}},
						},
					},
				},
			},
			contains: "credentials ([`mytestSecret`](#mytestsecret))",
		},
		{
			paramName:  "testSecret2",
			stepParams: []string{"testSecret2"},
			stepData: &config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testSecret2"},
						},
					},
				},
			},
			contains: "",
		},
	}

	for _, test := range tt {
		stepParameterNames = test.stepParams
		res := parameterFurtherInfo(test.paramName, test.stepData)
		stepParameterNames = []string{}
		if len(test.contains) == 0 {
			assert.Equalf(t, test.contains, res, fmt.Sprintf("param %v", test.paramName))
		} else {
			assert.Containsf(t, res, test.contains, fmt.Sprintf("param %v", test.paramName))
		}
		for _, notThere := range test.notContains {
			assert.NotContainsf(t, res, notThere, fmt.Sprintf("param %v", test.paramName))
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

func TestFormatDefault(t *testing.T) {
	tt := []struct {
		param      config.StepParameters
		stepParams []string
		expected   string
		contains   []string
	}{
		{param: config.StepParameters{Name: "test1"}, stepParams: []string{"test1"}, expected: "`$PIPER_test1` (if set)"},
		{param: config.StepParameters{Name: "jenkins1"}, stepParams: []string{}, expected: ""},
		{
			param:      config.StepParameters{Name: "test1", Default: []conditionDefault{{key: "key1", value: "val1", def: "def1"}, {key: "key2", value: "val2", def: "def2"}}},
			stepParams: []string{"test1"},
			contains:   []string{"key1=`val1`: `def1`", "key2=`val2`: `def2`"},
		},
		{
			param:      config.StepParameters{Name: "test1", Default: []interface{}{conditionDefault{key: "key1", value: "val1", def: "def1"}, "def2"}},
			stepParams: []string{"test1"},
			contains:   []string{"key1=`val1`: `def1`", "- `def2`"},
		},
		{
			param:      config.StepParameters{Name: "test1", Default: map[string]string{"key1": "def1", "key2": "def2"}},
			stepParams: []string{"test1"},
			contains:   []string{"`key1`: `def1`", "`key2`: `def2`"},
		},
		{param: config.StepParameters{Name: "test1", Default: ""}, stepParams: []string{"test1"}, expected: "`''`"},
		{param: config.StepParameters{Name: "test1", Default: "def1"}, stepParams: []string{"test1"}, expected: "`def1`"},
		{param: config.StepParameters{Name: "test1", Default: 1}, stepParams: []string{"test1"}, expected: "`1`"},
	}

	for _, test := range tt {
		if len(test.contains) > 0 {
			res := formatDefault(test.param, test.stepParams)
			for _, check := range test.contains {
				assert.Contains(t, res, check)
			}

		} else {
			assert.Equal(t, test.expected, formatDefault(test.param, test.stepParams))
		}

	}
}

func TestAliasList(t *testing.T) {
	tt := []struct {
		aliases  []config.Alias
		expected string
		contains []string
	}{
		{aliases: []config.Alias{}, expected: "-"},
		{aliases: []config.Alias{{Name: "alias1"}}, expected: "`alias1`"},
		{aliases: []config.Alias{{Name: "alias1", Deprecated: true}}, expected: "`alias1` (**deprecated**)"},
		{aliases: []config.Alias{{Name: "alias1"}, {Name: "alias2", Deprecated: true}}, contains: []string{"- `alias1`", "- `alias2` (**deprecated**)"}},
	}

	for _, test := range tt {
		if len(test.contains) > 0 {
			res := aliasList(test.aliases)
			for _, check := range test.contains {
				assert.Contains(t, res, check)
			}
		} else {
			assert.Equal(t, test.expected, aliasList(test.aliases))
		}
	}
}

func TestPossibleValueList(t *testing.T) {
	tt := []struct {
		possibleValues []interface{}
		expected       string
		contains       []string
	}{
		{possibleValues: []interface{}{}, expected: ""},
		{possibleValues: []interface{}{"poss1", 0}, contains: []string{"- `poss1`", "- `0`"}},
	}

	for _, test := range tt {
		if len(test.contains) > 0 {
			res := possibleValueList(test.possibleValues)
			for _, check := range test.contains {
				assert.Contains(t, res, check)
			}
		} else {
			assert.Equal(t, test.expected, possibleValueList(test.possibleValues))
		}
	}
}

func TestScopeDetails(t *testing.T) {
	tt := []struct {
		scope    []string
		contains []string
	}{
		{scope: []string{"PARAMETERS", "GENERAL", "STEPS", "STAGES"}, contains: []string{"<li>&#9746; parameter</li>", "<li>&#9746; general</li>", "<li>&#9746; steps</li>", "<li>&#9746; stages</li>"}},
		{scope: []string{}, contains: []string{"<li>&#9744; parameter</li>", "<li>&#9744; general</li>", "<li>&#9744; steps</li>", "<li>&#9744; stages</li>"}},
	}

	for _, test := range tt {
		res := scopeDetails(test.scope)

		for _, c := range test.contains {
			assert.Contains(t, res, c)
		}
	}

}

func TestResourceReferenceDetails(t *testing.T) {
	tt := []struct {
		resourceRef []config.ResourceReference
		expected    string
		contains    []string
	}{
		{resourceRef: []config.ResourceReference{}, expected: "none"},
		{
			resourceRef: []config.ResourceReference{
				{Name: "commonPipelineEnvironment", Aliases: []config.Alias{}, Type: "", Param: "testParam"},
			},
			expected: "_commonPipelineEnvironment_:<br />&nbsp;&nbsp;reference to: `testParam`<br />",
		},
		{
			resourceRef: []config.ResourceReference{
				{Name: "testCredentialId", Aliases: []config.Alias{}, Type: "secret", Param: "password"},
			},
			expected: "Jenkins credential id:<br />&nbsp;&nbsp;id: `testCredentialId`<br />&nbsp;&nbsp;reference to: `password`<br />",
		},
		{
			resourceRef: []config.ResourceReference{
				{Name: "testCredentialId", Aliases: []config.Alias{{Name: "alias1"}, {Name: "alias2", Deprecated: true}}, Type: "secret", Param: "password"},
			},
			contains: []string{"&nbsp;&nbsp;aliases:<br />", "&nbsp;&nbsp;- `alias1`<br />", "&nbsp;&nbsp;- `alias2` (**Deprecated**)<br />"},
		},
	}

	for _, test := range tt {
		if len(test.contains) > 0 {
			res := resourceReferenceDetails(test.resourceRef)
			for _, check := range test.contains {
				assert.Contains(t, res, check)
			}
		} else {
			assert.Equal(t, test.expected, resourceReferenceDetails(test.resourceRef))
		}
	}
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
