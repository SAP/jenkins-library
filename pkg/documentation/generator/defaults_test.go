package generator

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

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
