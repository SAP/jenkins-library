package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

type errReadCloser int

func (errReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (errReadCloser) Close() error {
	return nil
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

	testConfig := `general:
  p3: p3_general
  px3: px3_general
  p4: p4_general
steps:
  step1:
    p4: p4_step
    px4: px4_step
    p5: p5_step
stages:
  stage1:
    p5: p5_stage
    px5: px5_stage
    p6: p6_stage
`
	filters := StepFilters{
		General:    []string{"p0", "p1", "p2", "p3", "p4"},
		Steps:      []string{"p0", "p1", "p2", "p3", "p4", "p5"},
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
	stepConfig := c.GetStepConfig(flags, paramJSON, myConfig, defaults, filters, "stage1", "step1")

	t.Run("Config", func(t *testing.T) {
		expected := map[string]string{
			"p0": "p0_general_default",
			"p1": "p1_step_default",
			"p2": "p2_general_default",
			"p3": "p3_general",
			"p4": "p4_step",
			"p5": "p5_stage",
			"p6": "p6_param",
			"p7": "p7_flag",
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
			Source:         map[string]interface{}{"key1": "value1"},
			Filter:         []string{},
			MergeData:      map[string]interface{}{"key2": "value2"},
			ExpectedOutput: map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
		{
			Source:         map[string]interface{}{"key1": "value1"},
			Filter:         []string{"key1"},
			MergeData:      map[string]interface{}{"key2": "value2"},
			ExpectedOutput: map[string]interface{}{"key2": "value2"},
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
			if len(stepConfig.Config) != len(row.ExpectedOutput) {
				t.Errorf("Mixin  was incorrect, got: %v, expected: %v", stepConfig.Config, row.ExpectedOutput)
			}
		})
	}
}
