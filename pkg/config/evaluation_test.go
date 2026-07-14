//go:build unit
// +build unit

package config

import (
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
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
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
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "simple stepActive conditions",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
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
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "explicit active/deactivate over stepActiveCondition",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
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
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "stepNotActiveCondition over stepActiveCondition",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
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
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "stepNotActiveCondition over explicitly activated step",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
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
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "deactivate if only active step in stage",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
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
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "OnlyActiveStepInStage: one of the next steps is active",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
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
				},
			},
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

func TestRunConfigV1EvaluateConditionsV1Housekeeping(t *testing.T) {
	config := Config{Stages: map[string]map[string]interface{}{
		"Test Stage 1": {
			"housekeeping1": true,      // explicit activation of a housekeeping step
			"testKey":       "testVal", // condition for primary steps
		},
	}}
	envRootPath := ".pipeline"

	tests := []struct {
		name                 string
		pipelineConfig       PipelineDefinitionV1
		stagesWithExtensions []string
		files                []string
		githubActions        bool
		wantRunSteps         map[string]map[string]bool
		wantRunStages        map[string]bool
	}{
		{
			name: "two housekeeping steps active by default do not activate stage",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping2",
						Housekeeping: true,
					}, {
						Name:         "housekeeping3",
						Housekeeping: true,
					}, {
						Name:       "primary1",
						Conditions: []StepCondition{{ConfigKey: "notExistentKey"}},
					}},
				},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": false,
					"housekeeping3": false,
					"primary1":      false,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "housekeeping steps follow own conditions when a primary step is active",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping2",
						Housekeeping: true,
					}, {
						Name:         "housekeeping3",
						Housekeeping: true,
						Conditions:   []StepCondition{{ConfigKey: "notExistentKey"}},
					}, {
						Name:       "primary1",
						Conditions: []StepCondition{{ConfigKey: "testKey"}},
					}},
				},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": true,
					"housekeeping3": false,
					"primary1":      true,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "explicit activation of housekeeping step does not activate stage",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping1",
						Housekeeping: true,
					}, {
						Name:       "primary1",
						Conditions: []StepCondition{{ConfigKey: "notExistentKey"}},
					}},
				},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping1": false,
					"primary1":      false,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "stage with announced extension activates housekeeping-only stage",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping2",
						Housekeeping: true,
					}, {
						Name:         "housekeeping3",
						Housekeeping: true,
					}},
				},
			}}},
			stagesWithExtensions: []string{"Test Stage 1"},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": true,
					"housekeeping3": true,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "extension-only stage with zero active steps is active",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:       "primary1",
						Conditions: []StepCondition{{ConfigKey: "notExistentKey"}},
					}},
				},
			}}},
			stagesWithExtensions: []string{"Test Stage 1"},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"primary1": false,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "local post extension activates stage on GitHub Actions",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping2",
						Housekeeping: true,
					}},
				},
			}}},
			files:         []string{".pipeline/extensions/postTest Stage 1/action.yml"},
			githubActions: true,
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": true,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "local pre extension with yaml suffix activates stage on GitHub Actions",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping2",
						Housekeeping: true,
					}},
				},
			}}},
			files:         []string{".pipeline/extensions/preTest Stage 1/action.yaml"},
			githubActions: true,
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": true,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
		{
			name: "local extension is ignored outside GitHub Actions",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping2",
						Housekeeping: true,
					}},
				},
			}}},
			files:         []string{".pipeline/extensions/postTest Stage 1/action.yml"},
			githubActions: false,
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": false,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "local extension of another stage does not activate stage",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:         "housekeeping2",
						Housekeeping: true,
					}},
				},
			}}},
			files:         []string{".pipeline/extensions/postOther Stage/action.yml"},
			githubActions: true,
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": false,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "mixed legacy onlyActiveStepInStage and housekeeping without active primary",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:                "housekeeping2",
						Housekeeping:        true,
						NotActiveConditions: []StepCondition{{OnlyActiveStepInStage: true}},
					}, {
						Name:                "housekeeping3",
						Housekeeping:        true,
						NotActiveConditions: []StepCondition{{OnlyActiveStepInStage: true}},
					}, {
						Name:       "primary1",
						Conditions: []StepCondition{{ConfigKey: "notExistentKey"}},
					}},
				},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": false,
					"housekeeping3": false,
					"primary1":      false,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": false},
		},
		{
			name: "mixed legacy onlyActiveStepInStage and housekeeping with active primary",
			pipelineConfig: PipelineDefinitionV1{Spec: Spec{Stages: []Stage{
				{
					DisplayName: "Test Stage 1",
					Steps: []Step{{
						Name:                "housekeeping2",
						Housekeeping:        true,
						NotActiveConditions: []StepCondition{{OnlyActiveStepInStage: true}},
					}, {
						Name:                "housekeeping3",
						Housekeeping:        true,
						NotActiveConditions: []StepCondition{{OnlyActiveStepInStage: true}},
					}, {
						Name:       "primary1",
						Conditions: []StepCondition{{ConfigKey: "testKey"}},
					}},
				},
			}}},
			wantRunSteps: map[string]map[string]bool{
				"Test Stage 1": {
					"housekeeping2": true,
					"housekeeping3": true,
					"primary1":      true,
				},
			},
			wantRunStages: map[string]bool{"Test Stage 1": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.githubActions {
				t.Setenv("GITHUB_ACTIONS", "true")
			} else {
				// Make the test deterministic when running inside GitHub Actions CI.
				t.Setenv("GITHUB_ACTION", "false")
				t.Setenv("GITHUB_ACTIONS", "false")
			}
			filesMock := mock.FilesMock{}
			for _, file := range tt.files {
				filesMock.AddFile(file, []byte("content"))
			}
			r := &RunConfigV1{PipelineConfig: tt.pipelineConfig, StagesWithExtensions: tt.stagesWithExtensions}
			assert.NoError(t, r.evaluateConditionsV1(&config, &filesMock, envRootPath))

			assert.Equal(t, tt.wantRunSteps, r.RunSteps, "RunSteps mismatch")
			assert.Equal(t, tt.wantRunStages, r.RunStages, "RunStages mismatch")
		})
	}
}

type globErrorFilesMock struct {
	mock.FilesMock
}

func (g *globErrorFilesMock) Glob(pattern string) ([]string, error) {
	return nil, fmt.Errorf("glob failure")
}

func TestStageHasExtension(t *testing.T) {
	const gitHubActions = "GitHubActions"

	t.Run("announced via StagesWithExtensions, independent of orchestrator", func(t *testing.T) {
		r := &RunConfigV1{StagesWithExtensions: []string{"Acceptance", "Performance"}}
		assert.True(t, r.stageHasExtension("Acceptance", "Jenkins", &mock.FilesMock{}))
	})
	t.Run("local extension yml on GitHub Actions", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile(".pipeline/extensions/postAcceptance/action.yml", []byte("content"))
		r := &RunConfigV1{}
		assert.True(t, r.stageHasExtension("Acceptance", gitHubActions, &filesMock))
	})
	t.Run("local extension yaml on GitHub Actions", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile(".pipeline/extensions/preAcceptance/action.yaml", []byte("content"))
		r := &RunConfigV1{}
		assert.True(t, r.stageHasExtension("Acceptance", gitHubActions, &filesMock))
	})
	t.Run("no extension present", func(t *testing.T) {
		r := &RunConfigV1{}
		assert.False(t, r.stageHasExtension("Acceptance", gitHubActions, &mock.FilesMock{}))
	})
	t.Run("local extension ignored on other orchestrators", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile(".pipeline/extensions/postAcceptance/action.yml", []byte("content"))
		r := &RunConfigV1{}
		assert.False(t, r.stageHasExtension("Acceptance", "Jenkins", &filesMock))
	})
	t.Run("glob error is treated as no extension", func(t *testing.T) {
		r := &RunConfigV1{}
		assert.False(t, r.stageHasExtension("Acceptance", gitHubActions, &globErrorFilesMock{}))
	})
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
	err := os.MkdirAll(filepath.Join(cpeDir, "custom"), 0o700)
	if err != nil {
		t.Fatal("Failed to create sub directories")
	}
	os.WriteFile(filepath.Join(cpeDir, "myCpeTrueFile"), []byte("myTrueValue"), 0o700)
	os.WriteFile(filepath.Join(cpeDir, "custom", "myCpeTrueFile"), []byte("myTrueValue"), 0o700)

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

func TestCheckForNpmScriptsInPackagesV1(t *testing.T) {
	const validPkg = `{"scripts": {"build": "echo build"}}`
	const invalidPkg = `not a json`

	tests := []struct {
		name        string
		files       map[string]string
		npmScript   string
		config      StepConfig
		expected    bool
		expectError bool
	}{
		{
			name:      "node_modules excluded by default",
			files:     map[string]string{"node_modules/foo/package.json": validPkg},
			npmScript: "build",
			expected:  false,
		},
		{
			name:      "gen excluded by default",
			files:     map[string]string{"gen/foo/package.json": validPkg},
			npmScript: "build",
			expected:  false,
		},
		{
			name:      "tmp excluded by default",
			files:     map[string]string{"tmp/foo/package.json": validPkg},
			npmScript: "build",
			expected:  false,
		},
		{
			name:        "invalid package.json without exclude returns error",
			files:       map[string]string{"test_fixtures/package.json": invalidPkg},
			npmScript:   "build",
			expectError: true,
		},
		{
			name: "invalid package.json under custom exclude is skipped",
			files: map[string]string{
				"package.json":               validPkg,
				"test_fixtures/package.json": invalidPkg,
			},
			npmScript: "build",
			config: StepConfig{Config: map[string]any{
				"npmScriptsPackageJSONExcludes": []any{"test_fixtures/**"},
			}},
			expected: true,
		},
		{
			name: "custom exclude accepts []string",
			files: map[string]string{
				"package.json":               validPkg,
				"test_fixtures/package.json": invalidPkg,
			},
			npmScript: "build",
			config: StepConfig{Config: map[string]any{
				"npmScriptsPackageJSONExcludes": []string{"test_fixtures/**"},
			}},
			expected: true,
		},
		{
			name:      "missing script returns false",
			files:     map[string]string{"package.json": validPkg},
			npmScript: "nonexistent",
			expected:  false,
		},
		{
			name: "non-string entries in exclude list are ignored",
			files: map[string]string{
				"package.json":               validPkg,
				"test_fixtures/package.json": invalidPkg,
			},
			npmScript: "build",
			config: StepConfig{Config: map[string]any{
				"npmScriptsPackageJSONExcludes": []any{42, "test_fixtures/**", nil},
			}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := mock.FilesMock{}
			for path, content := range tt.files {
				files.AddFile(path, []byte(content))
			}

			result, err := checkForNpmScriptsInPackagesV1(tt.npmScript, tt.config, &files)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNpmScriptsPackageJSONExcludesFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   StepConfig
		expected []string
	}{
		{
			name:     "missing key returns nil",
			config:   StepConfig{Config: map[string]any{}},
			expected: nil,
		},
		{
			name:     "nil value returns nil",
			config:   StepConfig{Config: map[string]any{"npmScriptsPackageJSONExcludes": nil}},
			expected: nil,
		},
		{
			name: "[]string passthrough",
			config: StepConfig{Config: map[string]any{
				"npmScriptsPackageJSONExcludes": []string{"a/**", "b/**"},
			}},
			expected: []string{"a/**", "b/**"},
		},
		{
			name: "[]any of strings is converted",
			config: StepConfig{Config: map[string]any{
				"npmScriptsPackageJSONExcludes": []any{"a/**", "b/**"},
			}},
			expected: []string{"a/**", "b/**"},
		},
		{
			name: "empty strings dropped, non-strings ignored",
			config: StepConfig{Config: map[string]any{
				"npmScriptsPackageJSONExcludes": []any{"", "a/**", 7, nil, "b/**"},
			}},
			expected: []string{"a/**", "b/**"},
		},
		{
			name: "unsupported type returns nil",
			config: StepConfig{Config: map[string]any{
				"npmScriptsPackageJSONExcludes": "a/**",
			}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, npmScriptsPackageJSONExcludesFromConfig(tt.config))
		})
	}
}
