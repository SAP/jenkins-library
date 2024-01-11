//go:build unit
// +build unit

package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
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
		fileContent = io.NopCloser(strings.NewReader(`
		{
			"scripts": {
				"npmScript": "echo test",
				"npmScript2": "echo test"
			}
		}
		`))
	case "_package.json":
		fileContent = io.NopCloser(strings.NewReader("wrong json format"))
	case "test/package.json":
		fileContent = io.NopCloser(strings.NewReader("{}"))
	}
	return fileContent, nil
}

func TestRunConfigV1EvaluateConditionsV1(t *testing.T) {
	config := Config{Stages: map[string]map[string]interface{}{
		"Test Stage 1": {
			"step1":    true,       // explicit activate
			"step5":    true,       // explicit activate
			"step2":    false,      // explicit deactivate
			"testKey":  "testVal",  // some condition 1
			"testKey2": "testVal2", // some condition 2
		},
	}}
	filesMock := mock.FilesMock{}
	envRootPath := ".pipeline"

	tests := []struct {
		name           string
		pipelineConfig PipelineDefinitionV1
		wantRunSteps   map[string]map[string]bool
		wantRunStages  map[string]bool
	}{
		{
			name: "all steps in stage are inactive",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{{DisplayName: "Test Stage 1",
				Steps: []Step{{
					Name:                "step1",
					NotActiveConditions: []StepCondition{{ConfigKey: "testKey"}},
				}, {
					Name: "step2",
				}, {
					Name:                "step3",
					NotActiveConditions: []StepCondition{{ConfigKey: "testKey"}},
				}},
			},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"step1": false,
					"step2": false,
					"step3": false,
				}},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "simple stepActive conditions",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{{DisplayName: "Test Stage 1",
				Steps: []Step{{
					Name:       "step3",
					Conditions: []StepCondition{{ConfigKey: "testKey"}},
				}, {
					Name:       "step4",
					Conditions: []StepCondition{{ConfigKey: "notExistentKey"}},
				}},
			},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"step3": true,
					"step4": false,
				}},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "explicit active/deactivate over stepActiveCondition",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{{DisplayName: "Test Stage 1",
				Steps: []Step{{
					Name:       "step1",
					Conditions: []StepCondition{{ConfigKey: "notExistentKey"}},
				}, {
					Name:       "step2",
					Conditions: []StepCondition{{ConfigKey: "testKey"}},
				}},
			},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"step1": true,
					"step2": false,
				}},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "stepNotActiveCondition over stepActiveCondition",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{{DisplayName: "Test Stage 1",
				Steps: []Step{{
					Name:                "step3",
					Conditions:          []StepCondition{{ConfigKey: "testKey"}},
					NotActiveConditions: []StepCondition{{ConfigKey: "testKey2"}},
				}, {
					// false notActive condition
					Name:                "step4",
					Conditions:          []StepCondition{{ConfigKey: "testKey"}},
					NotActiveConditions: []StepCondition{{ConfigKey: "notExistentKey"}},
				}},
			},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"step3": false,
					"step4": true,
				}},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "stepNotActiveCondition over explicitly activated step",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{{DisplayName: "Test Stage 1",
				Steps: []Step{{
					Name:                "step1",
					NotActiveConditions: []StepCondition{{ConfigKey: "testKey"}},
				}, {
					Name:                "step5",
					NotActiveConditions: []StepCondition{{ConfigKey: "notExistentKey"}},
				}},
			},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"step1": false,
					"step5": true,
				}},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "deactivate if only active step in stage",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{{DisplayName: "Test Stage 1",
				Steps: []Step{{
					Name:                "step1",
					NotActiveConditions: []StepCondition{{ConfigKey: "testKey"}},
				}, {
					Name: "step2",
				}, {
					Name:                "step3",
					NotActiveConditions: []StepCondition{{OnlyActiveStepInStage: true}},
				}, {
					Name:       "step4",
					Conditions: []StepCondition{{ConfigKey: "keyNotExist"}},
				}},
			},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"step1": false,
					"step2": false,
					"step3": false,
					"step4": false,
				}},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "OnlyActiveStepInStage: one of the next steps is active",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{{DisplayName: "Test Stage 1",
				Steps: []Step{{
					Name:                "step1",
					NotActiveConditions: []StepCondition{{ConfigKey: "testKey"}},
				}, {
					Name: "step2",
				}, {
					Name:                "step3",
					Conditions:          []StepCondition{{ConfigKey: "testKey"}},
					NotActiveConditions: []StepCondition{{OnlyActiveStepInStage: true}},
				}, {
					Name:       "step4",
					Conditions: []StepCondition{{ConfigKey: "testKey2"}},
				}},
			},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"step1": false,
					"step2": false,
					"step3": true,
					"step4": true,
				}},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RunConfigV1{PipelineConfig: tt.pipelineConfig}
			assert.NoError(t, r.evaluateConditionsV1(&config, &filesMock, envRootPath),
				fmt.Sprintf("evaluateConditionsV1() err, pipelineConfig = %v", tt.pipelineConfig),
			)

			assert.Equal(t, tt.wantRunSteps, r.RunSteps, "RunSteps mismatch")
			assert.Equal(t, tt.wantRunStages, r.RunStages, "RunStages mismatch")
		})
	}
}

func TestEvaluateV1(t *testing.T) {
	tt := []struct {
		name          string
		config        StepConfig
		stepCondition StepCondition
		runSteps      map[string]bool
		expected      bool
		expectedError error
	}{
		{
			name: "Config condition - true",
			config: StepConfig{Config: map[string]interface{}{
				"deployTool": "helm3",
			}},
			stepCondition: StepCondition{Config: map[string][]interface{}{"deployTool": {"helm", "helm3", "kubectl"}}},
			expected:      true,
		},
		{
			name: "Config condition - false",
			config: StepConfig{Config: map[string]interface{}{
				"deployTool": "notsupported",
			}},
			stepCondition: StepCondition{Config: map[string][]interface{}{"deployTool": {"helm", "helm3", "kubectl"}}},
			expected:      false,
		},
		{
			name: "Config condition - integer - true",
			config: StepConfig{Config: map[string]interface{}{
				"executors": 1,
			}},
			stepCondition: StepCondition{Config: map[string][]interface{}{"executors": {1}}},
			expected:      true,
		},
		{
			name: "Config condition - wrong condition definition",
			config: StepConfig{Config: map[string]interface{}{
				"deployTool": "helm3",
			}},
			stepCondition: StepCondition{Config: map[string][]interface{}{"deployTool": {"helm", "helm3", "kubectl"}, "deployTool2": {"myTool"}}},
			expectedError: fmt.Errorf("only one config key allowed per condition but 2 provided"),
		},
		{
			name: "ConfigKey condition - true",
			config: StepConfig{Config: map[string]interface{}{
				"dockerRegistryUrl": "https://my.docker.registry.url",
			}},
			stepCondition: StepCondition{ConfigKey: "dockerRegistryUrl"},
			expected:      true,
		},
		{
			name:          "ConfigKey condition - false",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{ConfigKey: "dockerRegistryUrl"},
			expected:      false,
		},
		{
			name: "nested ConfigKey condition - true",
			config: StepConfig{Config: map[string]interface{}{
				"cloudFoundry": map[string]interface{}{"space": "dev"},
			}},
			stepCondition: StepCondition{ConfigKey: "cloudFoundry/space"},
			expected:      true,
		},
		{
			name: "nested ConfigKey condition - false",
			config: StepConfig{Config: map[string]interface{}{
				"cloudFoundry": map[string]interface{}{"noSpace": "dev"},
			}},
			stepCondition: StepCondition{ConfigKey: "cloudFoundry/space"},
			expected:      false,
		},
		{
			name:          "FilePattern condition - true",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{FilePattern: "**/conf.js"},
			expected:      true,
		},
		{
			name:          "FilePattern condition - false",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{FilePattern: "**/confx.js"},
			expected:      false,
		},
		{
			name: "FilePatternFromConfig condition - true",
			config: StepConfig{Config: map[string]interface{}{
				"newmanCollection": "**/*.postman_collection.json",
			}},
			stepCondition: StepCondition{FilePatternFromConfig: "newmanCollection"},
			expected:      true,
		},
		{
			name: "FilePatternFromConfig condition - false",
			config: StepConfig{Config: map[string]interface{}{
				"newmanCollection": "**/*.postmanx_collection.json",
			}},
			stepCondition: StepCondition{FilePatternFromConfig: "newmanCollection"},
			expected:      false,
		},
		{
			name: "FilePatternFromConfig condition - false, empty value",
			config: StepConfig{Config: map[string]interface{}{
				"newmanCollection": "",
			}},
			stepCondition: StepCondition{FilePatternFromConfig: "newmanCollection"},
			expected:      false,
		},
		{
			name:          "NpmScript condition - true",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{NpmScript: "testScript"},
			expected:      true,
		},
		{
			name:          "NpmScript condition - true",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{NpmScript: "missingScript"},
			expected:      false,
		},
		{
			name:          "Inactive condition - false",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{Inactive: true},
			expected:      false,
		},
		{
			name:          "Inactive condition - true",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{Inactive: false},
			expected:      true,
		},
		{
			name:          "CommonPipelineEnvironment - true",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{CommonPipelineEnvironment: map[string]interface{}{"myCpeTrueFile": "myTrueValue"}},
			expected:      true,
		},
		{
			name:          "CommonPipelineEnvironment - false",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{CommonPipelineEnvironment: map[string]interface{}{"myCpeTrueFile": "notMyTrueValue"}},
			expected:      false,
		},
		{
			name:          "CommonPipelineEnvironmentVariableExists - true",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{PipelineEnvironmentFilled: "custom/myCpeTrueFile"},
			expected:      true,
		},
		{
			name:          "CommonPipelineEnvironmentVariableExists - false",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{PipelineEnvironmentFilled: "custom/notMyCpeTrueFile"},
			expected:      false,
		},
		{
			name:          "NotActiveCondition: all previous steps are inactive",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{OnlyActiveStepInStage: true},
			runSteps:      map[string]bool{"step1": false, "step2": false},
			expected:      true,
		},
		{
			name:          "NotActiveCondition: one of the previous steps is active",
			config:        StepConfig{Config: map[string]interface{}{}},
			stepCondition: StepCondition{OnlyActiveStepInStage: true},
			runSteps:      map[string]bool{"step1": false, "step2": false, "step3": true},
			expected:      false,
		},
		{
			name:     "No condition - true",
			config:   StepConfig{Config: map[string]interface{}{}},
			expected: true,
		},
	}

	packageJson := `{
	"scripts": {
		"testScript": "whatever"
	}
}`

	filesMock := mock.FilesMock{}
	filesMock.AddFile("conf.js", []byte("//test"))
	filesMock.AddFile("my.postman_collection.json", []byte("{}"))
	filesMock.AddFile("package.json", []byte(packageJson))

	dir := t.TempDir()

	cpeDir := filepath.Join(dir, "commonPipelineEnvironment")
	err := os.MkdirAll(filepath.Join(cpeDir, "custom"), 0700)
	if err != nil {
		t.Fatal("Failed to create sub directories")
	}
	os.WriteFile(filepath.Join(cpeDir, "myCpeTrueFile"), []byte("myTrueValue"), 0700)
	os.WriteFile(filepath.Join(cpeDir, "custom", "myCpeTrueFile"), []byte("myTrueValue"), 0700)

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			active, err := test.stepCondition.evaluateV1(test.config, &filesMock, "dummy", dir, test.runSteps)
			if test.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, fmt.Sprint(test.expectedError))
			}
			assert.Equal(t, test.expected, active)
		})
	}
}

func TestEvaluateConditions(t *testing.T) {
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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
			stageConfig: io.NopCloser(strings.NewReader(`
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

func TestAnyOtherStepIsActive(t *testing.T) {
	targetStep := "step3"

	tests := []struct {
		name     string
		runSteps map[string]bool
		want     bool
	}{
		{
			name: "all steps are inactive (target active)",
			runSteps: map[string]bool{
				"step1": false,
				"step2": false,
				"step3": true,
				"step4": false,
			},
			want: false,
		},
		{
			name: "all steps are inactive (target inactive)",
			runSteps: map[string]bool{
				"step1": false,
				"step2": false,
				"step3": false,
				"step4": false,
			},
			want: false,
		},
		{
			name: "some previous step is active",
			runSteps: map[string]bool{
				"step1": false,
				"step2": true,
				"step3": false,
				"step4": false,
			},
			want: true,
		},
		{
			name: "some next step is active",
			runSteps: map[string]bool{
				"step1": false,
				"step2": false,
				"step3": true,
				"step4": true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, anyOtherStepIsActive(targetStep, tt.runSteps), "anyOtherStepIsActive(%v, %v)", targetStep, tt.runSteps)
		})
	}
}
