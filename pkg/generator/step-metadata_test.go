package main

import (
	//"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func configOpenFileMock(name string) (io.ReadCloser, error) {
	meta1 := `metadata:
  name: testStep
  description: Test description
  longDescription: |
    Long Test description
  spec:
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

func writeFileMock(filename string, data []byte, perm os.FileMode) error {
	return nil
}

func TestProcessMetaFiles(t *testing.T) {
	processMetaFiles([]string{"test.yaml"}, configOpenFileMock, writeFileMock)

	//ToDo: asserts!
}

func TestSetDefaultParameters(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		stepData := config.StepData{
			Spec: config.StepSpec{
				Inputs: config.StepInputs{
					Parameters: []config.StepParameters{
						{Name: "param0", Scope: []string{"GENERAL"}, Type: "string", Default: "val0"},
						{Name: "param1", Scope: []string{"STEPS"}, Type: "string"},
						{Name: "param2", Scope: []string{"STAGES"}, Type: "bool", Default: true},
						{Name: "param3", Scope: []string{"PARAMETERS"}, Type: "bool"},
						{Name: "param4", Scope: []string{"ENV"}, Type: "[]string", Default: []string{"val4_1", "val4_2"}},
						{Name: "param5", Scope: []string{"ENV"}, Type: "[]string"},
					},
				},
			},
		}

		expected := []string{
			"\"val0\"",
			"os.Getenv(\"PIPER_param1\")",
			"true",
			"false",
			"[]string{\"val4_1\", \"val4_2\"}",
			"[]string{}",
		}

		err := setDefaultParameters(&stepData)

		assert.NoError(t, err, "error occured but none expected")

		for k, v := range expected {
			assert.Equal(t, v, stepData.Spec.Inputs.Parameters[k].Default, fmt.Sprintf("default not correct for parameter %v", k))
		}
	})

	t.Run("error case", func(t *testing.T) {
		stepData := []config.StepData{
			{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "param0", Scope: []string{"GENERAL"}, Type: "int", Default: 10},
							{Name: "param1", Scope: []string{"GENERAL"}, Type: "int"},
						},
					},
				},
			},
			{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "param1", Scope: []string{"GENERAL"}, Type: "int"},
						},
					},
				},
			},
		}

		for k, v := range stepData {
			err := setDefaultParameters(&v)
			assert.Error(t, err, fmt.Sprintf("error expected but none occured for parameter %v", k))
		}
	})
}

func TestGetStepInfo(t *testing.T) {

	stepData := config.StepData{
		Metadata: config.StepMetadata{
			Name:            "testStep",
			Description:     "Test description",
			LongDescription: "Long Test description",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{Name: "param0", Scope: []string{"GENERAL"}, Type: "string", Default: "test"},
				},
			},
		},
	}

	myStepInfo := getStepInfo(&stepData)

	assert.Equal(t, "testStep", myStepInfo.StepName, "StepName incorrect")
	assert.Equal(t, "TestStepCommand", myStepInfo.CobraCmdFuncName, "CobraCmdFuncName incorrect")
	assert.Equal(t, "createTestStepCmd", myStepInfo.CreateCmdVar, "CreateCmdVar incorrect")
	assert.Equal(t, "Test description", myStepInfo.Short, "Short incorrect")
	assert.Equal(t, "Long Test description", myStepInfo.Long, "Long incorrect")
	assert.Equal(t, stepData.Spec.Inputs.Parameters, myStepInfo.Metadata, "Metadata incorrect")
	assert.Equal(t, "addTestStepFlags", myStepInfo.FlagsFunc, "FlagsFunc incorrect")

}

/*
func TestStepGeneration(t *testing.T) {
	var b bytes.Buffer
	g, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+".golden"))
	if err != nil {
		t.Fatalf("failed reading .golden: %s", err)
	}
	if !bytes.Equal(b.Bytes(), g) {
		t.Errorf("written json does not match .golden file")
	}
}
*/
