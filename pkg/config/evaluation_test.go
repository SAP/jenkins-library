package config

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func evaluateConditionsGlobMock(pattern string) ([]string, error) {
	matches := []string{}
	switch pattern {
	case "**/conf.js":
		matches = append(matches, "conf.js")
	case "**/package.json":
		matches = append(matches, "package.json", "package/node_modules/lib/package.json", "node_modules/package.json", "test/package.json")

	}
	return matches, nil
}

func evaluateConditionsOpenFileMock(name string, _ map[string]string) (io.ReadCloser, error) {
	var fileContent io.ReadCloser
	switch name {
	case "package.json":
		fileContent = ioutil.NopCloser(strings.NewReader(`
		{
			"scripts": {
				"npmScript": "echo test",
				"npmScript2": "echo test"
			}
		}
		`))
	case "_package.json":
		fileContent = ioutil.NopCloser(strings.NewReader("wrong json format"))
	case "test/package.json":
		fileContent = ioutil.NopCloser(strings.NewReader("{}"))
	}
	return fileContent, nil
}

func Test_evaluateConditions(t *testing.T) {
	tests := []struct {
		name             string
		customConfig     *Config
		stageConfig      io.ReadCloser
		runStepsExpected map[string]map[string]bool
		globFunc         func(pattern string) ([]string, error)
		wantErr          bool
	}{
		{
			name: "test config condition - success",
			customConfig: &Config{
				General: map[string]interface{}{
					"testGeneral": "myVal1",
				},
				Stages: map[string]map[string]interface{}{
					"testStage2": {
						"testStage": "myVal2",
					},
				},
				Steps: map[string]map[string]interface{}{
					"thirdStep": {
						"testStep1": "myVal3",
					},
				},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
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
        config: testStep1
      forthStep:
        config: testStep2
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {"firstStep": true},
				"testStage2": {"secondStep": true},
				"testStage3": {
					"thirdStep": true,
					"forthStep": false,
				},
			},
			wantErr: false,
		},
		{
			name: "test config condition - wrong usage with list",
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        config: 
         - testGeneral
            `)),
			runStepsExpected: map[string]map[string]bool{},
			wantErr:          true,
		},
		{
			name: "test config value condition - success",
			customConfig: &Config{
				General: map[string]interface{}{
					"testGeneral": "myVal1",
				},
				Stages: map[string]map[string]interface{}{},
				Steps: map[string]map[string]interface{}{
					"thirdStep": {
						"testStep": "myVal3",
					},
				},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        config:
          testGeneral:
            - myValx
            - myVal1
  testStage2:
    stepConditions:
      secondStep:
        config:
          testStage:
            - maValXyz
  testStage3:
    stepConditions:
      thirdStep:
        config:
          testStep:
            - myVal3
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {"firstStep": true},
				"testStage2": {"secondStep": false},
				"testStage3": {"thirdStep": true},
			},
			wantErr: false,
		},
		{
			name: "test configKey condition - success",
			customConfig: &Config{
				General: map[string]interface{}{
					"myKey1_1": "myVal1_1",
				},
				Stages: map[string]map[string]interface{}{},
				Steps: map[string]map[string]interface{}{
					"thirdStep": {
						"myKey3_1": "myVal3_1",
					},
				},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        configKeys:
          - myKey1_1
          - myKey1_2
  testStage2:
    stepConditions:
      secondStep:
        configKeys:
          - myKey2_1
  testStage3:
    stepConditions:
      thirdStep:
        configKeys:
          - myKey3_1
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {"firstStep": true},
				"testStage2": {"secondStep": false},
				"testStage3": {"thirdStep": true},
			},
			wantErr: false,
		},
		{
			name: "test configKey condition - not list",
			customConfig: &Config{
				General: map[string]interface{}{
					"myKey1_1": "myVal1_1",
				},
				Stages: map[string]map[string]interface{}{},
				Steps:  map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        configKeys: myKey1_1
            `)),
			wantErr: true,
		},
		{
			name: "test configKey condition - success",
			customConfig: &Config{
				General: map[string]interface{}{
					"myKey1_1": "myVal1_1",
				},
				Stages: map[string]map[string]interface{}{},
				Steps: map[string]map[string]interface{}{
					"thirdStep": {
						"myKey3_1": "myVal3_1",
					},
				},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        configKeys:
          - myKey1_1
          - myKey1_2
  testStage2:
    stepConditions:
      secondStep:
        configKeys:
          - myKey2_1
  testStage3:
    stepConditions:
      thirdStep:
        configKeys:
          - myKey3_1
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {"firstStep": true},
				"testStage2": {"secondStep": false},
				"testStage3": {"thirdStep": true},
			},
			wantErr: false,
		},
		{
			name:     "test filePattern condition - success",
			globFunc: evaluateConditionsGlobMock,
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePattern: '**/conf.js'
      secondStep:
        filePattern: '**/conf.jsx'
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep":  true,
					"secondStep": false,
				},
			},
			wantErr: false,
		},
		{
			name: "test filePattern condition - error while searching files by pattern",
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePattern: '**/conf.js'
            `)),
			runStepsExpected: map[string]map[string]bool{},
			globFunc: func(pattern string) ([]string, error) {
				return nil, errors.New("failed to check if file exists")
			},
			wantErr: true,
		},
		{
			name:     "test filePattern condition with list - success",
			globFunc: evaluateConditionsGlobMock,
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePattern:
         - '**/conf.js'
         - 'myCollection.json'
      secondStep:
        filePattern: 
         - '**/conf.jsx'
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep":  true,
					"secondStep": false,
				},
			},
			wantErr: false,
		},
		{
			name: "test filePattern condition with list - error while searching files by pattern",
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePattern:
         - '**/conf.js'
         - 'myCollection.json'
            `)),
			runStepsExpected: map[string]map[string]bool{},
			globFunc: func(pattern string) ([]string, error) {
				return nil, errors.New("failed to check if file exists")
			},
			wantErr: true,
		},
		{
			name:     "test filePatternFromConfig condition - success",
			globFunc: evaluateConditionsGlobMock,
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps: map[string]map[string]interface{}{
					"firstStep": {
						"myVal1": "**/conf.js",
					},
					"thirdStep": {
						"myVal3": "**/conf.jsx",
					},
				},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePatternFromConfig: myVal1
      secondStep:
        filePatternFromConfig: myVal2
      thirdStep:
        filePatternFromConfig: myVal3
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep":  true,
					"secondStep": false,
					"thirdStep":  false,
				},
			},
			wantErr: false,
		},
		{
			name: "test filePatternFromConfig condition - error while searching files by pattern",
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps: map[string]map[string]interface{}{
					"firstStep": {
						"myVal1": "**/conf.js",
					},
				},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePatternFromConfig: myVal1
            `)),
			runStepsExpected: map[string]map[string]bool{},
			globFunc: func(pattern string) ([]string, error) {
				return nil, errors.New("failed to check if file exists")
			},
			wantErr: true,
		},
		{
			name:     "test npmScripts condition - success",
			globFunc: evaluateConditionsGlobMock,
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        npmScripts: 'npmScript'
      secondStep:
        npmScripts: 'npmScript1'
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep":  true,
					"secondStep": false,
				},
			},
			wantErr: false,
		},
		{
			name:     "test npmScripts condition with list - success",
			globFunc: evaluateConditionsGlobMock,
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        npmScripts:
         - 'npmScript'
         - 'npmScript2'
      secondStep:
        npmScripts:
         - 'npmScript3'
         - 'npmScript4'
            `)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep":  true,
					"secondStep": false,
				},
			},
			wantErr: false,
		},
		{
			name: "test npmScripts condition - json with wrong format",
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        npmScripts:
         - 'npmScript3'
            `)),
			runStepsExpected: map[string]map[string]bool{},
			globFunc: func(pattern string) ([]string, error) {
				return []string{"_package.json"}, nil
			},
			wantErr: true,
		},
		{
			name: "test npmScripts condition - error while searching package.json",
			customConfig: &Config{
				General: map[string]interface{}{},
				Stages:  map[string]map[string]interface{}{},
				Steps:   map[string]map[string]interface{}{},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        npmScripts:
         - 'npmScript3'
            `)),
			runStepsExpected: map[string]map[string]bool{},
			globFunc: func(pattern string) ([]string, error) {
				return nil, errors.New("failed to check if file exists")
			},
			wantErr: true,
		},
		{
			name: "test explicit activation / de-activation of step",
			customConfig: &Config{
				Stages: map[string]map[string]interface{}{
					"testStage1": {
						"firstStep":    true,
						"fisecondStep": false,
					},
				},
			},
			stageConfig: ioutil.NopCloser(strings.NewReader(`
stages:
  testStage1:
    stepConditions:
      firstStep:
        config: testGeneral
      secondStep:
        config: testStage
`)),
			runStepsExpected: map[string]map[string]bool{
				"testStage1": {
					"firstStep":  true,
					"secondStep": false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runConfig := RunConfig{
				StageConfigFile: tt.stageConfig,
				RunSteps:        map[string]map[string]bool{},
				OpenFile:        evaluateConditionsOpenFileMock,
			}
			err := runConfig.loadConditions()
			assert.NoError(t, err)
			err = runConfig.evaluateConditions(tt.customConfig, nil, nil, nil, nil, tt.globFunc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.runStepsExpected, runConfig.RunSteps)
			}
		})
	}
}
