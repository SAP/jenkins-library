//go:build unit
// +build unit

package generator

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateParameterOverview(t *testing.T) {

	stepData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Resources: []config.StepResources{
					{Name: "testStash", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{Name: "param1"},
					{Name: "param2", Mandatory: true},
					{Name: "param3", MandatoryIf: []config.ParameterDependence{{Name: "param1", Value: "param3Necessary"}}},
					{Name: "dockerImage", Default: "testImage"},
					{Name: "stashContent", Default: "testStash"},
				},
			},
		},
	}
	stepParameterNames = []string{"param1", "param2", "param3"}

	t.Run("Test Step Section", func(t *testing.T) {

		expected := `| Name | Mandatory | Additional information |
| ---- | --------- | ---------------------- |
| [param1](#param1) | no |  |
| [param2](#param2) | **yes** |  |
| [param3](#param3) | **(yes)** | mandatory in case of:<br />- ` + "[`param1`](#param1)=`param3Necessary`" + ` |

`

		assert.Equal(t, expected, createParameterOverview(&stepData, false))
	})

	t.Run("Test Execution Section", func(t *testing.T) {
		expected := `| Name | Mandatory | Additional information |
| ---- | --------- | ---------------------- |
| [dockerImage](#dockerimage) | no |  |
| [stashContent](#stashcontent) | no | ![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen) |

`
		assert.Equal(t, expected, createParameterOverview(&stepData, true))
	})

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
			notContains: []string{"deprecated"},
			contains:    "id of credentials",
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
			notContains: []string{"deprecated"},
			contains:    "credentials ([`mytestSecret`](#mytestsecret))",
		},
		{
			paramName:  "testSecret1Deprecated",
			stepParams: []string{"testSecret1Deprecated"},
			stepData: &config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testSecret1Deprecated", Secret: true, ResourceRef: []config.ResourceReference{{Name: "mytestSecret", Type: "secret"}}, DeprecationMessage: "don't use"},
						},
					},
				},
			},
			contains: "![deprecated](https://img.shields.io/badge/-deprecated-red)![Secret](https://img.shields.io/badge/-Secret-yellowgreen)",
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
			notContains: []string{"deprecated"},
			contains:    "",
		},
		{
			paramName:  "testDeprecated",
			stepParams: []string{"testDeprecated"},
			stepData: &config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testDeprecated", DeprecationMessage: "don't use"},
						},
					},
				},
			},
			contains: "![deprecated](https://img.shields.io/badge/-deprecated-red)",
		},
	}

	for _, test := range tt {
		stepParameterNames = test.stepParams
		res, err := parameterFurtherInfo(test.paramName, test.stepData, false)
		assert.NoError(t, err)
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

func TestCheckParameterInfo(t *testing.T) {
	t.Parallel()
	tt := []struct {
		info                 string
		stepParam            bool
		executionEnvironment bool
		expected             string
		expectedErr          error
	}{
		{info: "step param", stepParam: true, executionEnvironment: false, expected: "step param", expectedErr: nil},
		{info: "execution param", stepParam: false, executionEnvironment: true, expected: "execution param", expectedErr: nil},
		{info: "step param in execution", stepParam: true, executionEnvironment: true, expected: "", expectedErr: fmt.Errorf("step parameter not relevant as execution environment parameter")},
		{info: "execution param in step", stepParam: false, executionEnvironment: false, expected: "", expectedErr: fmt.Errorf("execution environment parameter not relevant as step parameter")},
	}

	for _, test := range tt {
		result, err := checkParameterInfo(test.info, test.stepParam, test.executionEnvironment)
		if test.expectedErr == nil {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, fmt.Sprint(test.expectedErr))
		}
		assert.Equal(t, test.expected, result)
	}

}

func TestCreateParameterDetails(t *testing.T) {
	t.Run("default", func(t *testing.T) {
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
		assert.NotContains(t, res, "| Deprecated |")
	})

	t.Run("conditional mandatory parameters", func(t *testing.T) {
		stepData := config.StepData{
			Spec: config.StepSpec{
				Inputs: config.StepInputs{
					Parameters: []config.StepParameters{
						{
							Name:        "param2",
							MandatoryIf: []config.ParameterDependence{{Name: "param1", Value: "param1Val"}},
							Type:        "string",
						},
					},
				},
			},
		}

		res := createParameterDetails(&stepData)

		assert.Contains(t, res, "mandatory in case of:<br />- [`param1`](#param1)=`param1Val`")
	})

	t.Run("deprecated parameters", func(t *testing.T) {
		stepData := config.StepData{
			Spec: config.StepSpec{
				Inputs: config.StepInputs{
					Parameters: []config.StepParameters{
						{
							Name:               "param2",
							Type:               "string",
							DeprecationMessage: "this is deprecated",
						},
					},
				},
			},
		}

		res := createParameterDetails(&stepData)

		assert.Contains(t, res, "| Deprecated |")
		assert.Contains(t, res, "this is deprecated")
	})

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
			expected: "Jenkins credential id:<br />&nbsp;&nbsp;id: [`testCredentialId`](#testcredentialid)<br />&nbsp;&nbsp;reference to: `password`<br />",
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
					{Name: "ab7", MandatoryIf: []config.ParameterDependence{{Name: "ab1", Value: "ab1Val1"}}},
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
		assert.Equal(t, "ab7", stepData.Spec.Inputs.Parameters[6].Name)
	})

	t.Run("consider mandatory", func(t *testing.T) {
		sortStepParameters(&stepData, true)

		assert.Equal(t, "ab2", stepData.Spec.Inputs.Parameters[0].Name)
		assert.Equal(t, "ab3", stepData.Spec.Inputs.Parameters[1].Name)
		assert.Equal(t, "ab6", stepData.Spec.Inputs.Parameters[2].Name)
		assert.Equal(t, "ab7", stepData.Spec.Inputs.Parameters[3].Name)
		assert.Equal(t, "ab1", stepData.Spec.Inputs.Parameters[4].Name)
		assert.Equal(t, "ab4", stepData.Spec.Inputs.Parameters[5].Name)
		assert.Equal(t, "ab5", stepData.Spec.Inputs.Parameters[6].Name)
	})
}
