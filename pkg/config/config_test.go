package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

func customDefaultsOpenFileMock(name string) (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader("general:\n  p0: p0_custom_default")), nil
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
			},
		}

		stepMeta := StepData{Spec: StepSpec{Inputs: StepInputs{Parameters: parameterMetadata}}}

		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}

		// clean up tmp dir
		defer os.RemoveAll(dir)

		piperenv.SetParameter(filepath.Join(dir, "commonPipelineEnvironment"), "test_pe1", "pe1_val")

		stepAliases := []Alias{{Name: "stepAlias"}}
		stepConfig, err := c.GetStepConfig(flags, paramJSON, myConfig, defaults, filters, parameterMetadata, stepMeta.GetResourceParameters(dir, "commonPipelineEnvironment"), "stage1", "step1", stepAliases)

		assert.Equal(t, nil, err, "error occured but none expected")

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

	t.Run("Consider custom defaults from config", func(t *testing.T) {
		var c Config
		testConfDefaults := "customDefaults:\n- testDefaults.yaml"

		c.openFile = customDefaultsOpenFileMock

		stepConfig, err := c.GetStepConfig(nil, "", ioutil.NopCloser(strings.NewReader(testConfDefaults)), nil, StepFilters{General: []string{"p0"}}, nil, nil, "stage1", "step1", []Alias{})

		assert.NoError(t, err, "Error occured but no error expected")
		assert.Equal(t, "p0_custom_default", stepConfig.Config["p0"])

	})

	t.Run("Consider defaults from step config", func(t *testing.T) {
		var c Config

		stepParams := []StepParameters{StepParameters{Name: "p0", Scope: []string{"GENERAL"}, Type: "string", Default: "p0_step_default", Aliases: []Alias{{Name: "p0_alias"}}}}
		testConf := "general:\n p1: p1_conf"

		stepConfig, err := c.GetStepConfig(nil, "", ioutil.NopCloser(strings.NewReader(testConf)), nil, StepFilters{General: []string{"p0", "p1"}}, stepParams, nil, "stage1", "step1", []Alias{})

		assert.NoError(t, err, "Error occured but no error expected")
		assert.Equal(t, "p0_step_default", stepConfig.Config["p0"])
		assert.Equal(t, "p1_conf", stepConfig.Config["p1"])
	})

	t.Run("Failure case config", func(t *testing.T) {
		var c Config
		myConfig := ioutil.NopCloser(strings.NewReader("invalid config"))
		_, err := c.GetStepConfig(nil, "", myConfig, nil, StepFilters{}, []StepParameters{}, nil, "stage1", "step1", []Alias{})
		assert.EqualError(t, err, "failed to parse custom pipeline configuration: error unmarshalling \"invalid config\": error unmarshaling JSON: json: cannot unmarshal string into Go value of type config.Config", "default error expected")
	})

	t.Run("Failure case defaults", func(t *testing.T) {
		var c Config
		myConfig := ioutil.NopCloser(strings.NewReader(""))
		myDefaults := []io.ReadCloser{ioutil.NopCloser(strings.NewReader("invalid defaults"))}
		_, err := c.GetStepConfig(nil, "", myConfig, myDefaults, StepFilters{}, []StepParameters{}, nil, "stage1", "step1", []Alias{})
		assert.EqualError(t, err, "failed to parse pipeline default configuration: error unmarshalling \"invalid defaults\": error unmarshaling JSON: json: cannot unmarshal string into Go value of type config.Config", "default error expected")
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

	filters := StepFilters{
		General: []string{"p1", "p2"},
		Stages:  []string{"p4"},
		Steps:   []string{"p6", "p8"},
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
			"stage1": map[string]interface{}{
				"p3_notused": "p3_stage",
				"p4_alias":   "p4_stage",
			},
		},
		Steps: map[string]map[string]interface{}{
			"step1": map[string]interface{}{
				"p5_notused": "p5_step",
				"p6_alias":   "p6_step",
				"p7":         "p7_step",
			},
			"stepAlias1": map[string]interface{}{
				"p7":       "p7_stepAlias",
				"p8_alias": "p8_stepAlias",
				"p9":       "p9_stepAlias",
			},
		},
	}

	stepAliases := []Alias{{Name: "stepAlias1"}}

	c.ApplyAliasConfig(p, filters, "stage1", "step1", stepAliases)

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
				"step1": map[string]interface{}{
					"p1": "p1_step",
					"p2": "p2_step",
				},
				"stepAlias1": map[string]interface{}{
					"p2": "p2_stepAlias",
					"p3": "p3_stepAlias",
				},
				"stepAlias2": map[string]interface{}{
					"p3": "p3_stepAlias2",
					"p4": "p4_stepAlias2",
				},
			},
		}

		expected := Config{
			Steps: map[string]map[string]interface{}{
				"step1": map[string]interface{}{
					"p1": "p1_step",
					"p2": "p2_step",
					"p3": "p3_stepAlias",
					"p4": "p4_stepAlias2",
				},
				"stepAlias1": map[string]interface{}{
					"p2": "p2_stepAlias",
					"p3": "p3_stepAlias",
				},
				"stepAlias2": map[string]interface{}{
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
				"stepAlias1": map[string]interface{}{
					"p2": "p2_stepAlias",
				},
			},
		}

		expected := Config{
			Steps: map[string]map[string]interface{}{
				"step1": map[string]interface{}{
					"p2": "p2_stepAlias",
				},
				"stepAlias1": map[string]interface{}{
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
	}

	for _, row := range testTable {
		t.Run(fmt.Sprintf("Merging %v into %v", row.MergeData, row.Source), func(t *testing.T) {
			stepConfig := StepConfig{Config: row.Source}
			stepConfig.mixIn(row.MergeData, row.Filter)
			assert.Equal(t, row.ExpectedOutput, stepConfig.Config, "Mixin  was incorrect")
		})
	}
}
