package helper

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func configOpenFileMock(name string) (io.ReadCloser, error) {
	meta1 := `metadata:
  name: testStep
  aliases:
    - name: testStepAlias
      deprecated: true
  description: Test description
  longDescription: |
    Long Test description
spec:
  outputs:
    resources:
      - name: commonPipelineEnvironment
        type: piperEnvironment
        params:
          - name: artifactVersion
          - name: git/commitId
          - name: git/branch
      - name: influxTest
        type: influx
        params:
          - name: m1
            fields:
              - name: f1
            tags:
              - name: t1
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

var files map[string][]byte

func writeFileMock(filename string, data []byte, perm os.FileMode) error {
	if files == nil {
		files = make(map[string][]byte)
	}
	files[filename] = data
	return nil
}

func TestProcessMetaFiles(t *testing.T) {

	stepHelperData := StepHelperData{configOpenFileMock, writeFileMock, ""}
	docuHelperData := DocuHelperData{IsGenerateDocu: false}
	ProcessMetaFiles([]string{"test.yaml"}, stepHelperData, docuHelperData)

	t.Run("step code", func(t *testing.T) {
		goldenFilePath := filepath.Join("testdata", t.Name()+"_generated.golden")
		expected, err := ioutil.ReadFile(goldenFilePath)
		if err != nil {
			t.Fatalf("failed reading %v", goldenFilePath)
		}
		assert.Equal(t, expected, files["cmd/testStep_generated.go"])
		t.Log(string(files["cmd/testStep_generated.go"]))
	})

	t.Run("test code", func(t *testing.T) {
		goldenFilePath := filepath.Join("testdata", t.Name()+"_generated.golden")
		expected, err := ioutil.ReadFile(goldenFilePath)
		if err != nil {
			t.Fatalf("failed reading %v", goldenFilePath)
		}
		assert.Equal(t, expected, files["cmd/testStep_generated_test.go"])
	})

	t.Run("custom step code", func(t *testing.T) {
		stepHelperData = StepHelperData{configOpenFileMock, writeFileMock, "piperOsCmd"}
		ProcessMetaFiles([]string{"test.yaml"}, stepHelperData, docuHelperData)

		goldenFilePath := filepath.Join("testdata", t.Name()+"_generated.golden")
		expected, err := ioutil.ReadFile(goldenFilePath)
		if err != nil {
			t.Fatalf("failed reading %v", goldenFilePath)
		}
		assert.Equal(t, expected, files["cmd/testStep_generated.go"])
		t.Log(string(files["cmd/testStep_generated.go"]))
	})
}

func TestSetDefaultParameters(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		sliceVals := []string{"val4_1", "val4_2"}
		stringSliceDefault := make([]interface{}, len(sliceVals))
		for i, v := range sliceVals {
			stringSliceDefault[i] = v
		}
		stepData := config.StepData{
			Spec: config.StepSpec{
				Inputs: config.StepInputs{
					Parameters: []config.StepParameters{
						{Name: "param0", Type: "string", Default: "val0"},
						{Name: "param1", Type: "string"},
						{Name: "param2", Type: "bool", Default: true},
						{Name: "param3", Type: "bool"},
						{Name: "param4", Type: "[]string", Default: stringSliceDefault},
						{Name: "param5", Type: "[]string"},
						{Name: "param6", Type: "int"},
						{Name: "param7", Type: "int", Default: 1},
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
			"0",
			"1",
		}

		osImport, err := setDefaultParameters(&stepData)

		assert.NoError(t, err, "error occured but none expected")

		assert.Equal(t, true, osImport, "import of os package required")

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
							{Name: "param0", Type: "n/a", Default: 10},
							{Name: "param1", Type: "n/a"},
						},
					},
				},
			},
			{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "param1", Type: "n/a"},
						},
					},
				},
			},
		}

		for k, v := range stepData {
			_, err := setDefaultParameters(&v)
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

	myStepInfo, err := getStepInfo(&stepData, true, "")

	assert.NoError(t, err)

	assert.Equal(t, "testStep", myStepInfo.StepName, "StepName incorrect")
	assert.Equal(t, "TestStepCommand", myStepInfo.CobraCmdFuncName, "CobraCmdFuncName incorrect")
	assert.Equal(t, "createTestStepCmd", myStepInfo.CreateCmdVar, "CreateCmdVar incorrect")
	assert.Equal(t, "Test description", myStepInfo.Short, "Short incorrect")
	assert.Equal(t, "Long Test description", myStepInfo.Long, "Long incorrect")
	assert.Equal(t, stepData.Spec.Inputs.Parameters, myStepInfo.StepParameters, "Metadata incorrect")
	assert.Equal(t, "addTestStepFlags", myStepInfo.FlagsFunc, "FlagsFunc incorrect")
	assert.Equal(t, "addTestStepFlags", myStepInfo.FlagsFunc, "FlagsFunc incorrect")

}

func TestLongName(t *testing.T) {
	tt := []struct {
		input    string
		expected string
	}{
		{input: "my long name with no ticks", expected: "my long name with no ticks"},
		{input: "my long name with `ticks`", expected: "my long name with ` + \"`\" + `ticks` + \"`\" + `"},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, longName(v.input), fmt.Sprintf("wrong long name for run %v", k))
	}
}

func TestGolangNameTitle(t *testing.T) {
	tt := []struct {
		input    string
		expected string
	}{
		{input: "testApi", expected: "TestAPI"},
		{input: "apiTest", expected: "APITest"},
		{input: "testUrl", expected: "TestURL"},
		{input: "testId", expected: "TestID"},
		{input: "testJson", expected: "TestJSON"},
		{input: "jsonTest", expected: "JSONTest"},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, golangNameTitle(v.input), fmt.Sprintf("wrong golang name for run %v", k))
	}
}

func TestFlagType(t *testing.T) {
	tt := []struct {
		input    string
		expected string
	}{
		{input: "bool", expected: "BoolVar"},
		{input: "int", expected: "IntVar"},
		{input: "string", expected: "StringVar"},
		{input: "[]string", expected: "StringSliceVar"},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, flagType(v.input), fmt.Sprintf("wrong flag type for run %v", k))
	}
}

func TestGetStringSliceFromInterface(t *testing.T) {
	tt := []struct {
		input    interface{}
		expected []string
	}{
		{input: []interface{}{"Test", 2}, expected: []string{"Test", "2"}},
		{input: "Test", expected: []string{"Test"}},
	}

	for _, v := range tt {
		assert.Equal(t, v.expected, getStringSliceFromInterface(v.input), "interface conversion failed")
	}
}
