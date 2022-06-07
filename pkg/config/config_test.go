package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/stretchr/testify/assert"
)

type errReadCloser int

func (errReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (errReadCloser) Close() error {
	return nil
}

func customDefaultsOpenFileMock(name string, tokens map[string]string) (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader("general:\n  p0: p0_custom_default\nstages:\n  stage1:\n    p1: p1_custom_default")), nil
}

func TestReadConfig(t *testing.T) {

	var c Config

	t.Run("Success case", func(t *testing.T) {

		myConfig := strings.NewReader("general:\n  generalTestKey: generalTestValue\nsteps:\n  testStep:\n    testStepKey: testStepValue")

		err := c.ReadConfig(ioutil.NopCloser(myConfig)) // NopCloser "no-ops" the closing interface since strings do not need to be closed
		if err != nil {
			t.Errorf("Got error although no error expected: %v", err)
		}

		if c.General["generalTestKey"] != "generalTestValue" {
			t.Errorf("General config- got: %v, expected: %v", c.General["generalTestKey"], "generalTestValue")
		}

		if c.Steps["testStep"]["testStepKey"] != "testStepValue" {
			t.Errorf("Step config - got: %v, expected: %v", c.Steps["testStep"]["testStepKey"], "testStepValue")
		}
	})

	t.Run("Read failure", func(t *testing.T) {
		var rc errReadCloser
		err := c.ReadConfig(rc)
		if err == nil {
			t.Errorf("Got no error although error expected.")
		}
	})

	t.Run("Unmarshalling failure", func(t *testing.T) {
		myConfig := strings.NewReader("general:\n  generalTestKey: generalTestValue\nsteps:\n  testStep:\n\ttestStepKey: testStepValue")
		err := c.ReadConfig(ioutil.NopCloser(myConfig))
		if err == nil {
			t.Errorf("Got no error although error expected.")
		}
	})

}

func TestGetStepConfig(t *testing.T) {

	t.Run("Success case", func(t *testing.T) {

		testConfig := `general:
  p3: p3_general
  px3: px3_general
  p4: p4_general
steps:
  step1:
    p4: p4_step
    px4: px4_step
    p5: p5_step
    dependentParameter: dependentValue
  stepAlias:
    p8: p8_stepAlias
stages:
  stage1:
    p5: p5_stage
    px5: px5_stage
    p6: p6_stage
`
		filters := StepFilters{
			General:    []string{"p0", "p1", "p2", "p3", "p4"},
			Steps:      []string{"p0", "p1", "p2", "p3", "p4", "p5", "p8", "dependentParameter", "pd1", "dependentValue", "pd2"},
			Stages:     []string{"p0", "p1", "p2", "p3", "p4", "p5", "p6"},
			Parameters: []string{"p0", "p1", "p2", "p3", "p4", "p5", "p6", "p7"},
			Env:        []string{"p0", "p1", "p2", "p3", "p4", "p5"},
		}

		defaults1 := `general:
  p0: p0_general_default
  px0: px0_general_default
  p1: p1_general_default
steps:
  step1:
    p1: p1_step_default
    px1: px1_step_default
    p2: p2_step_default
    dependentValue:
      pd1: pd1_dependent_default
`

		defaults2 := `general:
  p2: p2_general_default
  px2: px2_general_default
  p3: p3_general_default
`

		paramJSON := `{"p6":"p6_param","p7":"p7_param"}`

		flags := map[string]interface{}{"p7": "p7_flag"}

		var c Config
		defaults := []io.ReadCloser{ioutil.NopCloser(strings.NewReader(defaults1)), ioutil.NopCloser(strings.NewReader(defaults2))}

		myConfig := ioutil.NopCloser(strings.NewReader(testConfig))

		parameterMetadata := []StepParameters{
			{
				Name:  "pd1",
				Scope: []string{"STEPS"},
				Conditions: []Condition{
					{
						Params: []Param{
							{Name: "dependentParameter", Value: "dependentValue"},
						},
					},
				},
			},
			{
				Name:    "pd2",
				Default: "pd2_metadata_default",
				Scope:   []string{"STEPS"},
				Conditions: []Condition{
					{
						Params: []Param{
							{Name: "dependentParameter", Value: "dependentValue"},
						},
					},
				},
			},
			{
				Name:        "pe1",
				Scope:       []string{"STEPS"},
				ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "test_pe1"}},
				Type:        "string",
			},
		}
		secretMetadata := []StepSecrets{
			{
				Name: "sd1",
				Type: "jenkins",
			},
		}

		stepAliases := []Alias{{Name: "stepAlias"}}

		stepMeta := StepData{
			Spec: StepSpec{
				Inputs: StepInputs{
					Parameters: parameterMetadata,
					Secrets:    secretMetadata,
				},
			},
			Metadata: StepMetadata{
				Aliases: stepAliases,
			},
		}

		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}

		// clean up tmp dir
		defer os.RemoveAll(dir)

		piperenv.SetParameter(filepath.Join(dir, "commonPipelineEnvironment"), "test_pe1", "pe1_val")

		stepConfig, err := c.GetStepConfig(flags, paramJSON, myConfig, defaults, false, filters, stepMeta, stepMeta.GetResourceParameters(dir, "commonPipelineEnvironment"), "stage1", "step1")

		assert.Equal(t, nil, err, "error occurred but none expected")

		t.Run("Config", func(t *testing.T) {
			expected := map[string]string{
				"p0":  "p0_general_default",
				"p1":  "p1_step_default",
				"p2":  "p2_general_default",
				"p3":  "p3_general",
				"p4":  "p4_step",
				"p5":  "p5_stage",
				"p6":  "p6_param",
				"p7":  "p7_flag",
				"p8":  "p8_stepAlias",
				"pd1": "pd1_dependent_default",
				"pd2": "pd2_metadata_default",
				"pe1": "pe1_val",
			}

			for k, v := range expected {
				t.Run(k, func(t *testing.T) {
					if stepConfig.Config[k] != v {
						t.Errorf("got: %v, expected: %v", stepConfig.Config[k], v)
					}
				})
			}
		})

		t.Run("Config not expected", func(t *testing.T) {
			notExpectedKeys := []string{"px0", "px1", "px2", "px3", "px4", "px5"}
			for _, p := range notExpectedKeys {
				t.Run(p, func(t *testing.T) {
					if stepConfig.Config[p] != nil {
						t.Errorf("unexpected: %v", p)
					}
				})
			}
		})
	})

	t.Run("Success case - environment nil", func(t *testing.T) {

		testConfig := ""
		filters := StepFilters{
			General: []string{"p0"},
		}

		defaults1 := `general:
  p0: p0_general_default
`
		var c Config
		defaults := []io.ReadCloser{ioutil.NopCloser(strings.NewReader(defaults1))}

		myConfig := ioutil.NopCloser(strings.NewReader(testConfig))

		stepMeta := StepData{Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{
			{Name: "p0", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "p0"}}},
		}}}}

		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}

		// clean up tmp dir
		defer os.RemoveAll(dir)

		stepConfig, err := c.GetStepConfig(map[string]interface{}{}, "", myConfig, defaults, false, filters, stepMeta, stepMeta.GetResourceParameters(dir, "commonPipelineEnvironment"), "stage1", "step1")

		assert.Equal(t, nil, err, "error occurred but none expected")

		assert.Equal(t, "p0_general_default", stepConfig.Config["p0"])
	})

	t.Run("Consider custom defaults from config", func(t *testing.T) {
		var c Config
		testConfDefaults := "customDefaults:\n- testDefaults.yaml"

		c.openFile = customDefaultsOpenFileMock

		stepConfig, err := c.GetStepConfig(nil, "", ioutil.NopCloser(strings.NewReader(testConfDefaults)), nil, false, StepFilters{General: []string{"p0"}}, StepData{}, nil, "stage1", "step1")

		assert.NoError(t, err, "Error occurred but no error expected")
		assert.Equal(t, "p0_custom_default", stepConfig.Config["p0"])
		assert.Equal(t, "p1_custom_default", stepConfig.Config["p1"])

	})

	t.Run("Don't consider custom defaults from config", func(t *testing.T) {
		var c Config
		testConfDefaults := "customDefaults:\n- testDefaults.yaml"

		c.openFile = customDefaultsOpenFileMock

		stepConfig, err := c.GetStepConfig(nil, "", ioutil.NopCloser(strings.NewReader(testConfDefaults)), nil, true, StepFilters{General: []string{"p0"}}, StepData{}, nil, "stage1", "step1")

		assert.NoError(t, err, "Error occurred but no error expected")
		assert.Equal(t, nil, stepConfig.Config["p0"])
		assert.Equal(t, nil, stepConfig.Config["p1"])

	})

	t.Run("Consider defaults from step config", func(t *testing.T) {
		var c Config

		stepParams := []StepParameters{{Name: "p0", Scope: []string{"GENERAL"}, Type: "string", Default: "p0_step_default", Aliases: []Alias{{Name: "p0_alias"}}}}
		metadata := StepData{
			Spec: StepSpec{
				Inputs: StepInputs{
					Parameters: stepParams,
				},
			},
		}
		testConf := "general:\n p1: p1_conf"

		stepConfig, err := c.GetStepConfig(nil, "", ioutil.NopCloser(strings.NewReader(testConf)), nil, false, StepFilters{General: []string{"p0", "p1"}}, metadata, nil, "stage1", "step1")

		assert.NoError(t, err, "Error occurred but no error expected")
		assert.Equal(t, "p0_step_default", stepConfig.Config["p0"])
		assert.Equal(t, "p1_conf", stepConfig.Config["p1"])
	})

	t.Run("Ignore alias if wrong type", func(t *testing.T) {
		var c Config

		stepParams := []StepParameters{
			{Name: "p0", Scope: []string{"GENERAL"}, Type: "bool", Aliases: []Alias{}},
			{Name: "p1", Scope: []string{"GENERAL"}, Type: "string", Aliases: []Alias{{Name: "p0/subParam"}}}}
		metadata := StepData{
			Spec: StepSpec{
				Inputs: StepInputs{
					Parameters: stepParams,
				},
			},
		}
		testConf := "general:\n p0: true"

		stepConfig, err := c.GetStepConfig(nil, "", ioutil.NopCloser(strings.NewReader(testConf)), nil, false, StepFilters{General: []string{"p0", "p1"}}, metadata, nil, "stage1", "step1")

		assert.NoError(t, err, "Error occurred but no error expected")
		assert.Equal(t, true, stepConfig.Config["p0"])
		assert.Equal(t, nil, stepConfig.Config["p1"])
	})

	t.Run("Apply alias to paramJSON", func(t *testing.T) {
		var c Config

		secrets := []StepSecrets{
			{Name: "p0", Type: "string", Aliases: []Alias{{Name: "p1/subParam"}}}}
		metadata := StepData{
			Spec: StepSpec{
				Inputs: StepInputs{
					Secrets: secrets,
				},
			},
		}
		testConf := ""

		paramJSON := "{\"p1\":{\"subParam\":\"p1_value\"}}"
		stepConfig, err := c.GetStepConfig(nil, paramJSON, ioutil.NopCloser(strings.NewReader(testConf)), nil, true, StepFilters{Parameters: []string{"p0"}}, metadata, nil, "stage1", "step1")

		assert.NoError(t, err, "Error occurred but no error expected")
		assert.Equal(t, "p1_value", stepConfig.Config["p0"])
	})

	t.Run("Failure case config", func(t *testing.T) {
		var c Config
		myConfig := ioutil.NopCloser(strings.NewReader("invalid config"))
		_, err := c.GetStepConfig(nil, "", myConfig, nil, false, StepFilters{}, StepData{}, nil, "stage1", "step1")
		assert.EqualError(t, err, "failed to parse custom pipeline configuration: format of configuration is invalid \"invalid config\": error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type config.Config", "default error expected")
	})

	t.Run("Failure case defaults", func(t *testing.T) {
		var c Config
		myConfig := ioutil.NopCloser(strings.NewReader(""))
		myDefaults := []io.ReadCloser{ioutil.NopCloser(strings.NewReader("invalid defaults"))}
		_, err := c.GetStepConfig(nil, "", myConfig, myDefaults, false, StepFilters{}, StepData{}, nil, "stage1", "step1")
		assert.EqualError(t, err, "failed to read default configuration: error unmarshalling \"invalid defaults\": error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type config.Config", "default error expected")
	})

	t.Run("Test reporting parameters with aliases and cpe resources", func(t *testing.T) {
		var c Config
		testConfig := ioutil.NopCloser(strings.NewReader(`general:
  gcpJsonKeyFilePath: gcpJsonKeyFilePath_value
steps:
  step1:
    jsonKeyFilePath: gcpJsonKeyFilePath_from_alias`))
		testDefaults := []io.ReadCloser{ioutil.NopCloser(strings.NewReader(`general:
  pipelineId: gcsBucketId_from_alias
steps:
  step1:
    gcsBucketId: gcsBucketId_value`))}
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}

		// clean up tmp dir
		defer os.RemoveAll(dir)
		cpeDir := filepath.Join(dir, "commonPipelineEnvironment/custom")
		err = os.MkdirAll(cpeDir, 0700)
		if err != nil {
			t.Fatal("Failed to create sub directory")
		}

		err = ioutil.WriteFile(filepath.Join(cpeDir, "gcsFolderPath.json"), []byte("\"value_from_cpe\""), 0700)
		assert.NoError(t, err)

		stepMeta := StepData{Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{}}}}
		stepConfig, err := c.GetStepConfig(nil, "", testConfig, testDefaults, false, StepFilters{General: []string{"p0", "p1"}}, stepMeta, ReportingParameters.GetResourceParameters(dir, "commonPipelineEnvironment"), "stage1", "step1")

		assert.NoError(t, err, "Error occurred but no error expected")
		assert.Equal(t, "gcpJsonKeyFilePath_from_alias", stepConfig.Config["gcpJsonKeyFilePath"])
		assert.Equal(t, "gcsBucketId_value", stepConfig.Config["gcsBucketId"])
		assert.Equal(t, "value_from_cpe", stepConfig.Config["gcsFolderPath"])
	})

	//ToDo: test merging of env and parameters/flags
}

func TestGetStepConfigWithJSON(t *testing.T) {

	filters := StepFilters{All: []string{"key1"}}

	t.Run("Without flags", func(t *testing.T) {
		sc := GetStepConfigWithJSON(nil, `"key1":"value1","key2":"value2"`, filters)

		if sc.Config["key1"] != "value1" && sc.Config["key2"] == "value2" {
			t.Errorf("got: %v, expected: %v", sc.Config, StepConfig{Config: map[string]interface{}{"key1": "value1"}})
		}
	})

	t.Run("With flags", func(t *testing.T) {
		flags := map[string]interface{}{"key1": "flagVal1"}
		sc := GetStepConfigWithJSON(flags, `"key1":"value1","key2":"value2"`, filters)
		if sc.Config["key1"] != "flagVal1" {
			t.Errorf("got: %v, expected: %v", sc.Config["key1"], "flagVal1")
		}
	})
}

func TestGetStageConfig(t *testing.T) {

	testConfig := `general:
  p1: p1_general
  px1: px1_general
stages:
  stage1:
    p2: p2_stage
    px2: px2_stage
`
	defaults1 := `general:
  p0: p0_general_default
  px0: px0_general_default
`
	paramJSON := `{"p3":"p3_param"}`

	t.Run("Success case - with filters", func(t *testing.T) {

		acceptedParams := []string{"p0", "p1", "p2", "p3"}

		var c Config
		defaults := []io.ReadCloser{ioutil.NopCloser(strings.NewReader(defaults1))}

		myConfig := ioutil.NopCloser(strings.NewReader(testConfig))

		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}

		// clean up tmp dir
		defer os.RemoveAll(dir)

		stepConfig, err := c.GetStageConfig(paramJSON, myConfig, defaults, false, acceptedParams, "stage1")

		assert.Equal(t, nil, err, "error occurred but none expected")

		t.Run("Config", func(t *testing.T) {
			expected := map[string]string{
				"p0": "p0_general_default",
				"p1": "p1_general",
				"p2": "p2_stage",
				"p3": "p3_param",
			}

			for k, v := range expected {
				t.Run(k, func(t *testing.T) {
					if stepConfig.Config[k] != v {
						t.Errorf("got: %v, expected: %v", stepConfig.Config[k], v)
					}
				})
			}
		})

		t.Run("Config not expected", func(t *testing.T) {
			notExpectedKeys := []string{"px0", "px1", "px2"}
			for _, p := range notExpectedKeys {
				t.Run(p, func(t *testing.T) {
					if stepConfig.Config[p] != nil {
						t.Errorf("unexpected: %v", p)
					}
				})
			}
		})
	})

	t.Run("Success case - no filters", func(t *testing.T) {

		acceptedParams := []string{}

		var c Config
		defaults := []io.ReadCloser{ioutil.NopCloser(strings.NewReader(defaults1))}

		myConfig := ioutil.NopCloser(strings.NewReader(testConfig))

		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}

		// clean up tmp dir
		defer os.RemoveAll(dir)

		stepConfig, err := c.GetStageConfig(paramJSON, myConfig, defaults, false, acceptedParams, "stage1")

		assert.Equal(t, nil, err, "error occurred but none expected")

		t.Run("Config", func(t *testing.T) {
			expected := map[string]string{
				"p0":  "p0_general_default",
				"px0": "px0_general_default",
				"p1":  "p1_general",
				"px1": "px1_general",
				"p2":  "p2_stage",
				"px2": "px2_stage",
				"p3":  "p3_param",
			}

			for k, v := range expected {
				t.Run(k, func(t *testing.T) {
					if stepConfig.Config[k] != v {
						t.Errorf("got: %v, expected: %v", stepConfig.Config[k], v)
					}
				})
			}
		})
	})
}

func TestApplyAliasConfig(t *testing.T) {
	p := []StepParameters{
		{
			Name: "p0",
			Aliases: []Alias{
				{Name: "p0_notused"},
			},
		},
		{
			Name: "p1",
			Aliases: []Alias{
				{Name: "p1_alias"},
			},
		},
		{
			Name: "p2",
			Aliases: []Alias{
				{Name: "p2_alias/deep/test"},
			},
		},
		{
			Name: "p3",
			Aliases: []Alias{
				{Name: "p3_notused"},
			},
		},
		{
			Name: "p4",
			Aliases: []Alias{
				{Name: "p4_alias"},
				{Name: "p4_2nd_alias"},
			},
		},
		{
			Name: "p5",
			Aliases: []Alias{
				{Name: "p5_notused"},
			},
		},
		{
			Name: "p6",
			Aliases: []Alias{
				{Name: "p6_1st_alias"},
				{Name: "p6_alias"},
			},
		},
		{
			Name: "p7",
			Aliases: []Alias{
				{Name: "p7_alias"},
			},
		},
		{
			Name: "p8",
			Aliases: []Alias{
				{Name: "p8_alias"},
			},
		},
		{
			Name: "p9",
		},
	}
	s := []StepSecrets{
		{
			Name: "s1",
			Aliases: []Alias{
				{Name: "s1_alias"},
			},
		},
	}

	filters := StepFilters{
		General: []string{"p1", "p2"},
		Stages:  []string{"p4"},
		Steps:   []string{"p6", "p8", "s1"},
	}

	c := Config{
		General: map[string]interface{}{
			"p0_notused": "p0_general",
			"p1_alias":   "p1_general",
			"p2_alias": map[string]interface{}{
				"deep": map[string]interface{}{
					"test": "p2_general",
				},
			},
		},
		Stages: map[string]map[string]interface{}{
			"stage1": {
				"p3_notused": "p3_stage",
				"p4_alias":   "p4_stage",
			},
		},
		Steps: map[string]map[string]interface{}{
			"step1": {
				"p5_notused": "p5_step",
				"p6_alias":   "p6_step",
				"p7":         "p7_step",
			},
			"stepAlias1": {
				"p7":       "p7_stepAlias",
				"p8_alias": "p8_stepAlias",
				"p9":       "p9_stepAlias",
				"s1_alias": "s1_stepAlias",
			},
		},
	}

	stepAliases := []Alias{{Name: "stepAlias1"}}

	c.ApplyAliasConfig(p, s, filters, "stage1", "step1", stepAliases)

	t.Run("Global", func(t *testing.T) {
		assert.Nil(t, c.General["p0"])
		assert.Equal(t, "p1_general", c.General["p1"])
		assert.Equal(t, "p2_general", c.General["p2"])
	})

	t.Run("Stage", func(t *testing.T) {
		assert.Nil(t, c.General["p3"])
		assert.Equal(t, "p4_stage", c.Stages["stage1"]["p4"])
	})

	t.Run("Steps", func(t *testing.T) {
		assert.Nil(t, c.General["p5"])
		assert.Equal(t, "p6_step", c.Steps["step1"]["p6"])
		assert.Equal(t, "p7_step", c.Steps["step1"]["p7"])
		assert.Equal(t, "p8_stepAlias", c.Steps["step1"]["p8"])
		assert.Equal(t, "p9_stepAlias", c.Steps["step1"]["p9"])
		assert.Equal(t, "s1_stepAlias", c.Steps["step1"]["s1"])
	})

}

func TestGetDeepAliasValue(t *testing.T) {
	c := map[string]interface{}{
		"p0": "p0_val",
		"p1": 11,
		"p2": map[string]interface{}{
			"p2_0": "p2_0_val",
			"p2_1": map[string]interface{}{
				"p2_1_0": "p2_1_0_val",
			},
		},
	}
	tt := []struct {
		key      string
		expected interface{}
	}{
		{key: "p0", expected: "p0_val"},
		{key: "p1", expected: 11},
		{key: "p2/p2_0", expected: "p2_0_val"},
		{key: "p2/p2_1/p2_1_0", expected: "p2_1_0_val"},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, getDeepAliasValue(c, v.key), fmt.Sprintf("wrong return value for run %v", k+1))
	}
}

func TestCopyStepAliasConfig(t *testing.T) {
	t.Run("Step config available", func(t *testing.T) {
		c := Config{
			Steps: map[string]map[string]interface{}{
				"step1": {
					"p1": "p1_step",
					"p2": "p2_step",
				},
				"stepAlias1": {
					"p2": "p2_stepAlias",
					"p3": "p3_stepAlias",
				},
				"stepAlias2": {
					"p3": "p3_stepAlias2",
					"p4": "p4_stepAlias2",
				},
			},
		}

		expected := Config{
			Steps: map[string]map[string]interface{}{
				"step1": {
					"p1": "p1_step",
					"p2": "p2_step",
					"p3": "p3_stepAlias",
					"p4": "p4_stepAlias2",
				},
				"stepAlias1": {
					"p2": "p2_stepAlias",
					"p3": "p3_stepAlias",
				},
				"stepAlias2": {
					"p3": "p3_stepAlias2",
					"p4": "p4_stepAlias2",
				},
			},
		}

		c.copyStepAliasConfig("step1", []Alias{{Name: "stepAlias1"}, {Name: "stepAlias2"}})
		assert.Equal(t, expected, c)
	})

	t.Run("Step config not available", func(t *testing.T) {
		c := Config{
			Steps: map[string]map[string]interface{}{
				"stepAlias1": {
					"p2": "p2_stepAlias",
				},
			},
		}

		expected := Config{
			Steps: map[string]map[string]interface{}{
				"step1": {
					"p2": "p2_stepAlias",
				},
				"stepAlias1": {
					"p2": "p2_stepAlias",
				},
			},
		}

		c.copyStepAliasConfig("step1", []Alias{{Name: "stepAlias1"}})
		assert.Equal(t, expected, c)
	})
}

func TestGetJSON(t *testing.T) {

	t.Run("Success case", func(t *testing.T) {
		custom := map[string]interface{}{"key1": "value1"}
		json, err := GetJSON(custom)
		if err != nil {
			t.Errorf("Got error although no error expected: %v", err)
		}

		if json != `{"key1":"value1"}` {
			t.Errorf("got: %v, expected: %v", json, `{"key1":"value1"}`)
		}

	})
	t.Run("Marshalling failure", func(t *testing.T) {
		_, err := GetJSON(make(chan int))
		if err == nil {
			t.Errorf("Got no error although error expected")
		}
	})
}

func TestMerge(t *testing.T) {

	testTable := []struct {
		Source         map[string]interface{}
		Filter         []string
		MergeData      map[string]interface{}
		ExpectedOutput map[string]interface{}
	}{
		{
			Source:         map[string]interface{}{"key1": "baseValue"},
			Filter:         []string{},
			MergeData:      map[string]interface{}{"key1": "overwrittenValue"},
			ExpectedOutput: map[string]interface{}{"key1": "overwrittenValue"},
		},
		{
			Source:         map[string]interface{}{"key1": "value1"},
			Filter:         []string{},
			MergeData:      map[string]interface{}{"key2": "value2"},
			ExpectedOutput: map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
		{
			Source:         map[string]interface{}{"key1": "value1"},
			Filter:         []string{"key1"},
			MergeData:      map[string]interface{}{"key2": "value2"},
			ExpectedOutput: map[string]interface{}{"key1": "value1"},
		},
		{
			Source:         map[string]interface{}{"key1": map[string]interface{}{"key1_1": "value1"}},
			Filter:         []string{},
			MergeData:      map[string]interface{}{"key1": map[string]interface{}{"key1_2": "value2"}},
			ExpectedOutput: map[string]interface{}{"key1": map[string]interface{}{"key1_1": "value1", "key1_2": "value2"}},
		},
		{
			Source:         map[string]interface{}{"key1": "value1"},
			Filter:         []string{"key1", ".+Key$"},
			MergeData:      map[string]interface{}{"regexKey": "value2", "regexKeyIgnored": "value3", "Key": "value3"},
			ExpectedOutput: map[string]interface{}{"key1": "value1", "regexKey": "value2"},
		},
	}

	for _, row := range testTable {
		t.Run(fmt.Sprintf("Merging %v into %v", row.MergeData, row.Source), func(t *testing.T) {
			stepConfig := StepConfig{Config: row.Source}
			stepConfig.mixIn(row.MergeData, row.Filter)
			assert.Equal(t, row.ExpectedOutput, stepConfig.Config, "Mixin was incorrect")
		})
	}
}

func TestStepConfig_mixInHookConfig(t *testing.T) {
	type fields struct {
		Config     map[string]interface{}
		HookConfig map[string]interface{}
	}
	type args struct {
		mergeData map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{name: "Splunk only",
			fields: fields{
				Config:     nil,
				HookConfig: nil,
			},
			args: args{mergeData: map[string]interface{}{
				"splunk": map[string]interface{}{
					"dsn":      "dsn",
					"token":    "token",
					"sendLogs": "false",
				},
			}},
			want: map[string]interface{}{
				"splunk": map[string]interface{}{
					"dsn":      "dsn",
					"token":    "token",
					"sendLogs": "false",
				},
			},
		},
		{name: "Sentry only",
			fields: fields{
				Config:     nil,
				HookConfig: nil,
			},
			args: args{mergeData: map[string]interface{}{
				"sentry": map[string]interface{}{
					"dsn": "sentrydsn",
				},
			}},
			want: map[string]interface{}{
				"sentry": map[string]interface{}{
					"dsn": "sentrydsn",
				},
			},
		},
		{name: "ANS only",
			fields: fields{
				Config:     nil,
				HookConfig: nil,
			},
			args: args{mergeData: map[string]interface{}{
				"ans": map[string]interface{}{
					"serviceKey": "serviceKey",
				},
			}},
			want: map[string]interface{}{
				"ans": map[string]interface{}{
					"serviceKey": "serviceKey",
				},
			},
		},
		{name: "ANS, Splunk and Sentry",
			fields: fields{
				Config:     nil,
				HookConfig: nil,
			},
			args: args{mergeData: map[string]interface{}{
				"ans": map[string]interface{}{
					"serviceKey": "serviceKey",
				},
				"splunk": map[string]interface{}{
					"dsn":      "dsn",
					"token":    "token",
					"sendLogs": "false",
				},
				"sentry": map[string]interface{}{
					"dsn": "sentrydsn",
				},
			}},
			want: map[string]interface{}{
				"splunk": map[string]interface{}{
					"dsn":      "dsn",
					"token":    "token",
					"sendLogs": "false",
				},
				"sentry": map[string]interface{}{
					"dsn": "sentrydsn",
				},
				"ans": map[string]interface{}{
					"serviceKey": "serviceKey",
				},
			},
		},
		{name: "No Hook",
			fields: fields{
				Config:     nil,
				HookConfig: nil,
			},
			args: args{mergeData: nil},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StepConfig{
				Config:     tt.fields.Config,
				HookConfig: tt.fields.HookConfig,
			}
			s.mixInHookConfig(tt.args.mergeData)
			if !reflect.DeepEqual(s.HookConfig, tt.want) {
				t.Errorf("mixInHookConfig() = %v, want %v", s.HookConfig, tt.want)
			}
		})
	}
}

func TestMixInStepDefaults(t *testing.T) {
	tt := []struct {
		name       string
		stepConfig *StepConfig
		stepParams []StepParameters
		expected   map[string]interface{}
	}{
		{name: "empty", stepConfig: &StepConfig{}, stepParams: []StepParameters{}, expected: map[string]interface{}{}},
		{name: "no condition", stepConfig: &StepConfig{}, stepParams: []StepParameters{{Name: "noCondition", Default: "noCondition_default"}}, expected: map[string]interface{}{"noCondition": "noCondition_default"}},
		{
			name:       "with multiple conditions",
			stepConfig: &StepConfig{},
			stepParams: []StepParameters{
				{Name: "dependentParam1", Default: "dependentParam1_value"},
				{Name: "dependentParam2", Default: "dependentParam2_value"},
				{
					Name:    "withConditionParameter",
					Default: "withCondition_default_a",
					Conditions: []Condition{
						{ConditionRef: "strings-equal", Params: []Param{{Name: "dependentParam1", Value: "dependentParam1_value1"}}},
						{ConditionRef: "strings-equal", Params: []Param{{Name: "dependentParam2", Value: "dependentParam2_value1"}}},
					},
				},
				{
					Name:    "withConditionParameter",
					Default: "withCondition_default_b",
					Conditions: []Condition{
						{ConditionRef: "strings-equal", Params: []Param{{Name: "dependentParam1", Value: "dependentParam1_value2"}}},
						{ConditionRef: "strings-equal", Params: []Param{{Name: "dependentParam2", Value: "dependentParam2_value2"}}},
					},
				},
			},
			expected: map[string]interface{}{
				"dependentParam1":        "dependentParam1_value",
				"dependentParam2":        "dependentParam2_value",
				"dependentParam1_value1": map[string]interface{}{"withConditionParameter": "withCondition_default_a"},
				"dependentParam2_value1": map[string]interface{}{"withConditionParameter": "withCondition_default_a"},
				"dependentParam1_value2": map[string]interface{}{"withConditionParameter": "withCondition_default_b"},
				"dependentParam2_value2": map[string]interface{}{"withConditionParameter": "withCondition_default_b"},
			},
		},
	}

	for _, test := range tt {
		test.stepConfig.mixInStepDefaults(test.stepParams)
		assert.Equal(t, test.expected, test.stepConfig.Config, test.name)
	}
}

func TestCloneConfig(t *testing.T) {
	testConfig := &Config{
		General: map[string]interface{}{
			"p0": "p0_general",
		},
		Stages: map[string]map[string]interface{}{
			"stage1": {
				"p1": "p1_stage",
			},
		},
		Steps: map[string]map[string]interface{}{
			"step1": {
				"p2": "p2_step",
			},
		},
	}
	clone, err := cloneConfig(testConfig)
	assert.NoError(t, err)
	assert.Equal(t, testConfig, clone)
	testConfig.General["p0"] = "new_value"
	assert.NotEqual(t, testConfig.General, clone.General)
}
