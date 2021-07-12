package config

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type GetConfigMock struct {
	stepConfig map[string]map[string]StepConfig
	errors     map[string]error

	// passed parameters for checking
	configuration io.ReadCloser
	defaults      []io.ReadCloser
	filters       map[string]StepFilters
	parameters    map[string][]StepParameters
	secrets       map[string][]StepSecrets
	stageName     string
	stepName      string
	stepAliases   map[string][]Alias

	// not supported params will also not be handled in test
	flagValues    map[string]interface{}
	paramJSON     string
	envParameters map[string]interface{}
}

func (g *GetConfigMock) GetStepConfig(flagValues map[string]interface{}, paramJSON string, configuration io.ReadCloser, defaults []io.ReadCloser, ignoreCustomDefaults bool, filters StepFilters, parameters []StepParameters, secrets []StepSecrets, envParameters map[string]interface{}, stageName string, stepName string, stepAliases []Alias) (StepConfig, error) {

	if g.filters == nil {
		g.filters = map[string]StepFilters{}
	}
	if g.parameters == nil {
		g.parameters = map[string][]StepParameters{}
	}
	if g.secrets == nil {
		g.secrets = map[string][]StepSecrets{}
	}
	if g.stepAliases == nil {
		g.stepAliases = map[string][]Alias{}
	}

	g.configuration = configuration
	g.defaults = defaults
	g.filters[stepName] = filters
	g.parameters[stepName] = parameters
	g.secrets[stepName] = secrets
	g.stepAliases[stepName] = stepAliases

	if g.errors[stepName] != nil {
		return StepConfig{}, nil
	}
	return g.stepConfig[stageName][stepName], nil
}

func TestInitRunConfig(t *testing.T) {

	testData := struct {
		filters      map[string]StepFilters
		parameters   map[string][]StepParameters
		secrets      map[string][]StepSecrets
		stepAliases  map[string][]Alias
		stagesConfig map[string]map[string]interface{}
	}{
		filters: map[string]StepFilters{
			"firstStep": {All: []string{"filter"}},
		},
		parameters: map[string][]StepParameters{
			"firstStep": {{Name: "param"}},
		},
		secrets: map[string][]StepSecrets{
			"firstStep": {{Name: "secret"}},
		},
		stepAliases: map[string][]Alias{
			"firstStep": {{Name: "alias"}},
		},
	}

	t.Run("RunConfig E2E: test general step deactivation", func(t *testing.T) {
		runConfig := &RunConfig{
			Conditions: RunConditions{
				// StageConditions: map[string]StepConditions{
				StageConditions: map[string]map[string]PipelineConditions{
					"testStage1": {
						"firstStep": {
							Conditions: []v1beta1.PipelineTaskCondition{
								{
									ConditionRef: "config-equals",
									Params: []v1beta1.Param{
										{Name: "configKey", Value: *v1beta1.NewArrayOrString("testGeneral")},
										{Name: "contains", Value: *v1beta1.NewArrayOrString("myVal1")},
									},
								},
							},
						},
					},
					"testStage2": {
						"secondStep": {
							Conditions: []v1beta1.PipelineTaskCondition{
								{
									ConditionRef: "config-equals",
									Params: []v1beta1.Param{
										{Name: "configKey", Value: *v1beta1.NewArrayOrString("secondStepConfig")},
										{Name: "contains", Value: *v1beta1.NewArrayOrString("myValXyz")},
									},
								},
							},
						},
					},
					"testStage3": {
						"thirdStep": {
							Conditions: []v1beta1.PipelineTaskCondition{
								{
									ConditionRef: "config-equals",
									Params: []v1beta1.Param{
										{Name: "configKey", Value: *v1beta1.NewArrayOrString("testStep")},
										{Name: "contains", Value: *v1beta1.NewArrayOrString("myVal3")},
									},
								},
							},
						},
					},
				},
			},
		}

		config := Config{
			General: map[string]interface{}{
				"testGeneral": "myVal1",
			},
			Stages: map[string]map[string]interface{}{},
			Steps: map[string]map[string]interface{}{
				"thirdStep": {"testStep": "myVal3"},
			},
		}

		err := runConfig.InitRunConfig(&config, testData.stagesConfig, testData.filters, testData.parameters, testData.secrets, testData.stepAliases, nil)

		assert.NoError(t, err)

		// refactor stage deactivation
		// assert.True(t, runConfig.DeactivateStage["testStage2"])
		// assert.False(t, runConfig.DeactivateStage["testStage1"])
		// assert.False(t, runConfig.DeactivateStage["testStage3"])

		assert.True(t, runConfig.DeactivateStageSteps["testStage2"]["secondStep"])
		assert.False(t, runConfig.DeactivateStageSteps["testStage1"]["firstStep"])
		assert.False(t, runConfig.DeactivateStageSteps["testStage3"]["thirdStep"])
	})

	t.Run("RunConfig E2E: test evaluation of release pipeline stages", func(t *testing.T) {
		runConfig := &RunConfig{
			Conditions: RunConditions{
				// StageConditions: map[string]StepConditions{
				StageConditions: map[string]map[string]PipelineConditions{
					"setVersion": {
						"piperSetVersion": {
							Conditions: []v1beta1.PipelineTaskCondition{
								{
									ConditionRef: "config-equals",
									Params: []v1beta1.Param{
										{Name: "configKey", Value: *v1beta1.NewArrayOrString("version")},
										{Name: "contains", Value: *v1beta1.NewArrayOrString("v1.0.0")},
									},
								},
							},
						},
					},
					"build": {
						"piperMavenBuild": {
							Conditions: []v1beta1.PipelineTaskCondition{
								{
									ConditionRef: "file-exists",
									Params: []v1beta1.Param{
										{Name: "filePatternFromConfig", Value: *v1beta1.NewArrayOrString("piperMavenBuildPath")},
									},
								},
							},
						},
					},
					"postBuild": {
						"postBuildStep": {
							Conditions: []v1beta1.PipelineTaskCondition{
								{
									ConditionRef: "config-equals",
									Params: []v1beta1.Param{
										{Name: "configKey", Value: *v1beta1.NewArrayOrString("postBuildStepConfig")},
										{Name: "contains", Value: *v1beta1.NewArrayOrString("postBuildVal")},
									},
								},
							},
						},
					},
				},
			},
		}

		config := Config{
			General: map[string]interface{}{
				// "myKey1": "myVal1",
			},
			Stages: map[string]map[string]interface{}{},
			Steps: map[string]map[string]interface{}{
				"piperSetVersion": {"version": "v1.0.0"},
				"piperMavenBuild": {"piperMavenBuildPath": "**/pom.xml"},
				"postBuildStep":   {"postBuildStepConfig": "postBuildVal"},
			},
		}

		err := runConfig.InitRunConfig(&config, testData.stagesConfig, testData.filters, testData.parameters, testData.secrets, testData.stepAliases, nil)

		assert.NoError(t, err)

		// refactor stage deactivation
		// assert.True(t, runConfig.DeactivateStage["testStage2"])
		// assert.False(t, runConfig.DeactivateStage["testStage1"])
		// assert.False(t, runConfig.DeactivateStage["testStage3"])

		assert.True(t, runConfig.DeactivateStageSteps["build"]["piperMavenBuild"])
		assert.False(t, runConfig.DeactivateStageSteps["testStage1"]["firstStep"])
		assert.False(t, runConfig.DeactivateStageSteps["testStage3"]["thirdStep"])
	})

}

func TestRunConfigLoadConditions(t *testing.T) {

	stageConfigFile := "stage_config.yml"
	stageConfigContent := `stages:
  'testStage1':
    firstStep:
      conditions:
      - conditionRef: file-exists
        params:
        - name: filePattern
          value: '**/my.file'
`

	workingDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}
	defer os.RemoveAll(workingDir)

	t.Run("load conditions - file does not exists", func(t *testing.T) {
		runConfig := &RunConfig{ConditionFilePath: "file_does_not_exists"}
		err = runConfig.loadConditions()
		assert.Error(t, err, "error reading ")
	})

	t.Run("load conditions - file of invalid format", func(t *testing.T) {
		stageConfigFilePath := filepath.Join(workingDir, "test_file.txt")
		ioutil.WriteFile(stageConfigFilePath, []byte("-- {{ \\ wrong } file format }"), 0755)
		assert.FileExists(t, stageConfigFilePath)

		runConfig := &RunConfig{ConditionFilePath: stageConfigFilePath}
		err = runConfig.loadConditions()
		assert.Error(t, err, "format of configuration is invalid")
	})

	t.Run("load conditions - success", func(t *testing.T) {
		stageConfigFilePath := filepath.Join(workingDir, stageConfigFile)
		ioutil.WriteFile(stageConfigFilePath, []byte(stageConfigContent), 0755)
		assert.FileExists(t, stageConfigFilePath)

		runConfig := &RunConfig{ConditionFilePath: stageConfigFilePath}

		err = runConfig.loadConditions()

		condition := v1beta1.PipelineTaskCondition{
			ConditionRef: "file-exists",
			Params:       []v1beta1.Param{{Name: "filePattern", Value: *v1beta1.NewArrayOrString("**/my.file")}},
		}
		assert.NoError(t, err)
		assert.Equal(t, 1, len(runConfig.Conditions.StageConditions))
		assert.Equal(t, 1, len(runConfig.Conditions.StageConditions["testStage1"]))
		assert.Equal(t, condition, runConfig.Conditions.StageConditions["testStage1"]["firstStep"].Conditions[0])
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
