//go:build unit
// +build unit

package config

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func initRunConfigGlobMock(pattern string) ([]string, error) {
	matches := []string{}
	switch pattern {
	case "**/file1":
		matches = append(matches, "file1")
	case "directory/file2":
		matches = append(matches, "file2")
	}
	return matches, nil
}

func TestInitRunConfigV1(t *testing.T) {
	tt := []struct {
		name              string
		config            Config
		stageConfig       string
		runStagesExpected map[string]bool
		runStepsExpected  map[string]map[string]bool
		expectedError     error
		errorContains     string
	}{
		{
			name:             "success",
			config:           Config{Stages: map[string]map[string]interface{}{"testStage": {"testKey": "testVal"}}},
			stageConfig:      "spec:\n  stages:\n  - name: testStage\n    displayName: testStage\n    steps:\n    - name: testStep\n      conditions:\n      - configKey: testKey",
			runStepsExpected: map[string]map[string]bool{},
		},
		{
			name:             "error - load conditions",
			stageConfig:      "wrong stage config format",
			runStepsExpected: map[string]map[string]bool{},
			errorContains:    "failed to load pipeline run conditions",
		},
		{
			name:             "error - evaluate conditions",
			config:           Config{Stages: map[string]map[string]interface{}{"testStage": {"testKey": "testVal"}}},
			runStepsExpected: map[string]map[string]bool{},
			stageConfig:      "spec:\n  stages:\n  - name: testStage\n    displayName: testStage\n    steps:\n    - name: testStep\n      conditions:\n      - config:\n          configKey1:\n          - configVal1\n          configKey2:\n          - configVal2",
			errorContains:    "failed to evaluate step conditions",
		},
	}

	filesMock := mock.FilesMock{}

	for _, test := range tt {
		stageConfig := io.NopCloser(strings.NewReader(test.stageConfig))
		runConfig := RunConfig{StageConfigFile: stageConfig}
		runConfigV1 := RunConfigV1{RunConfig: runConfig}
		err := runConfigV1.InitRunConfigV1(&test.config, &filesMock, ".pipeline")
		if len(test.errorContains) > 0 {
			assert.Contains(t, fmt.Sprint(err), test.errorContains)
		} else {
			assert.NoError(t, err)
		}

	}
}

func TestInitRunConfig(t *testing.T) {
	tests := []struct {
		name             string
		customConfig     io.ReadCloser
		stageConfig      io.ReadCloser
		runStepsExpected map[string]map[string]bool
		wantErr          bool
	}{
		{
			name: "init run config with config condition - success",
			customConfig: io.NopCloser(strings.NewReader(`
general: 
  testGeneral: 'myVal1'
stages: 
  testStage2: 
    testStage: 'myVal2'
steps: 
  thirdStep: 
    testStep: 'myVal3'
            `)),
			stageConfig: io.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        config: testGeneral
  testStage2:
    stepConditions:
      secondStep:
        config: testStage
  testStage3:
    stepConditions:
      thirdStep:
        config: testStep
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep": true,
				},
				"testStage2": {
					"secondStep": true,
				},
				"testStage3": {
					"thirdStep": true,
				},
			},
			wantErr: false,
		},
		{
			name: "init run config with filePattern condition - success",
			customConfig: io.NopCloser(strings.NewReader(`
general: 
  testGeneral: 'myVal1'
stages: 
  testStage2: 
    testStage: 'myVal2'
steps: 
  thirdStep: 
    testStep: 'myVal3'
            `)),
			stageConfig: io.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePattern: "**/file1"
  testStage2:
    stepConditions:
      secondStep:
        filePattern: "directory/file2"
  testStage3:
    stepConditions:
      thirdStep:
        filePattern: "file3"
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep": true,
				},
				"testStage2": {
					"secondStep": true,
				},
				"testStage3": {
					"thirdStep": false,
				},
			},
			wantErr: false,
		},
		{
			name: "init run config - unknown condition in stage config",
			customConfig: io.NopCloser(strings.NewReader(`
steps: 
  testStep: 
    testConfig: 'testVal'
            `)),
			stageConfig: io.NopCloser(strings.NewReader(`
stages:
  testStage:
    stepConditions:
      testStep:
        wrongCondition: "condVal"
            `)),
			runStepsExpected: map[string]map[string]bool{},
			wantErr:          true,
		},
		{
			name:             "init run config - load conditions with invalid format",
			stageConfig:      io.NopCloser(strings.NewReader("wrong stage config format")),
			runStepsExpected: map[string]map[string]bool{},
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runConfig := RunConfig{StageConfigFile: tt.stageConfig}
			filter := StepFilters{All: []string{}, General: []string{}, Stages: []string{}, Steps: []string{}, Env: []string{}}
			projectConfig := Config{}
			_, err := projectConfig.GetStepConfig(map[string]interface{}{}, "", tt.customConfig,
				[]io.ReadCloser{}, false, filter, StepData{}, nil, "", "")
			assert.NoError(t, err)
			err = runConfig.InitRunConfig(&projectConfig, nil, nil, nil, nil, initRunConfigGlobMock, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.runStepsExpected, runConfig.RunSteps)
			}
		})
	}
}

func TestRunConfigLoadConditions(t *testing.T) {
	stageConfigContent := `stages:
  'testStage1':
    stepConditions:
      firstStep:
        filePattern: '**/my.file'
`
	t.Run("load conditions - file of invalid format", func(t *testing.T) {
		runConfig := &RunConfig{StageConfigFile: io.NopCloser(strings.NewReader("-- {{ \\ wrong } file format }"))}
		err := runConfig.loadConditions()
		assert.Error(t, err, "format of configuration is invalid")
	})

	t.Run("load conditions - success", func(t *testing.T) {
		runConfig := &RunConfig{StageConfigFile: io.NopCloser(strings.NewReader(stageConfigContent))}

		err := runConfig.loadConditions()
		assert.NoError(t, err)
		condition := map[string]interface{}{
			"filePattern": "**/my.file",
		}

		assert.Equal(t, 1, len(runConfig.StageConfig.Stages))
		assert.Equal(t, 1, len(runConfig.StageConfig.Stages["testStage1"].Conditions))
		assert.Equal(t, condition, runConfig.StageConfig.Stages["testStage1"].Conditions["firstStep"])
	})
}

func Test_stepConfigLookup(t *testing.T) {

	testConfig := map[string]interface{}{
		"general": map[string]interface{}{
			"generalKey": "generalValue",
		},
		"stages": map[string]interface{}{
			"testStep": map[string]interface{}{
				"stagesKey": "stagesValue",
			},
		},
		"steps": map[string]interface{}{
			"testStep": map[string]interface{}{
				"stepKey":            "stepValue",
				"stepKeyStringSlice": []string{"val1", "val2"},
			},
		},
		"configKey": "configValue",
	}

	type args struct {
		m        map[string]interface{}
		stepName string
		key      string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "empty map",
			args: args{nil, "", ""},
			want: nil,
		},
		{
			name: "key not in map, invalid stepName",
			args: args{testConfig, "some step", "some key"},
			want: nil,
		},
		{
			name: "key not in map, valid stepName",
			args: args{testConfig, "testStep", "some key"},
			want: nil,
		},
		{
			name: "key in map under general",
			args: args{testConfig, "some step", "generalKey"},
			want: "generalValue",
		},
		{
			name: "key in map under stages",
			args: args{testConfig, "testStep", "stagesKey"},
			want: "stagesValue",
		},
		{
			name: "key in map under general",
			args: args{testConfig, "testStep", "stepKey"},
			want: "stepValue",
		},
		{
			name: "key in map under general",
			args: args{testConfig, "testStep", "stepKeyStringSlice"},
			want: []string{"val1", "val2"},
		},
		{
			name: "key in map on top level string",
			args: args{testConfig, "", "configKey"},
			want: "configValue",
		},
		{
			name: "key in map on top level map",
			args: args{testConfig, "", "general"},
			want: map[string]interface{}{"generalKey": "generalValue"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stepConfigLookup(tt.args.m, tt.args.stepName, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stepConfigLookup() = %v, want %v", got, tt.want)
			}
		})
	}
}
