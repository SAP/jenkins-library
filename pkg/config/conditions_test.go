package config

import (
	"testing"

	"github.com/go-errors/errors"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func Test_validateCondition(t *testing.T) {
	type args struct {
		condition  v1beta1.PipelineTaskCondition
		stepConfig StepConfig
		stepName   string
	}
	tests := []struct {
		name          string
		args          args
		wantVaidation bool
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "unknown conditionRef (empty)",
			args: args{
				condition: v1beta1.PipelineTaskCondition{},
				stepName:  "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("validateCondition error: unknown conditionRef found for testStep").Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVaidation, err := validateCondition(tt.args.condition, tt.args.stepConfig, tt.args.stepName)
			if err != nil {
				if tt.wantErr {
					if err.Error() != tt.wantErrMsg {
						t.Errorf("validateCondition() errorMsg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
						return
					}
				} else {
					t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVaidation != tt.wantVaidation {
				t.Errorf("validateCondition() = %v, want %v", gotVaidation, tt.wantVaidation)
			}
		})
	}
}

func Test_validateCondition_configEquals(t *testing.T) {

	var testConfig = make(map[string]interface{})
	testConfig["emptyVal"] = ""
	testConfig["configEquals"] = "equalsTrue"
	testConfig["configExists"] = "configExistsValue"
	testConfig["filePatternKey"] = "**/conditions_test.go"
	testConfig["notStringVal"] = Config{}

	type args struct {
		condition  v1beta1.PipelineTaskCondition
		stepConfig StepConfig
		stepName   string
	}
	tests := []struct {
		name          string
		args          args
		wantVaidation bool
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "conditionRefConfigEquals wrong number of params",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation for testStep has unsupported number of string equal parameters (2 required, got 0)").Error(),
		},
		{
			name: "conditionRefConfigEquals wrong param definition",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
					Params: []v1beta1.Param{
						{Name: "", Value: *v1beta1.NewArrayOrString("myKey1", "myKey2")},
						{Name: "", Value: *v1beta1.NewArrayOrString("myKey")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation config-equals for testStep or has unsupported param definition or type").Error(),
		},
		{
			name: "conditionRefConfigEquals wrong param names",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
					Params: []v1beta1.Param{
						{Name: "random string1", Value: *v1beta1.NewArrayOrString("myKey1")},
						{Name: "contains", Value: *v1beta1.NewArrayOrString("myCompareValue")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation config-equals for testStep has unsupported param names").Error(),
		},
		{
			name: "conditionRefConfigEquals config value not a string",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("notStringVal")},
						{Name: "contains", Value: *v1beta1.NewArrayOrString("notStringVal")},
					},
				},
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("validateConfigEquals error: config value of notStringVal to compare with is not a string").Error(),
		},
		{
			name: "conditionRefConfigEquals config value not in map",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("notInMapKey")},
						{Name: "contains", Value: *v1beta1.NewArrayOrString("doesnt matter")},
					},
				},
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: false,
			wantErr:       false,
		},
		{
			name: "conditionRefConfigEquals config value empty / not equal",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("emptyVal")},
						{Name: "contains", Value: *v1beta1.NewArrayOrString("some other val")},
					},
				},
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: false,
			wantErr:       false,
		},
		{
			name: "conditionRefConfigEquals config value equals true",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("configEquals")},
						{Name: "contains", Value: *v1beta1.NewArrayOrString("equalsTrue")},
					},
				},
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: true,
			wantErr:       false,
		},
		{
			name: "conditionRefConfigEquals config value from array equals true",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-equals",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("configEquals")},
						{Name: "contains", Value: *v1beta1.NewArrayOrString("otherCompareVal", "equalsTrue")},
					},
				},
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVaidation, err := validateCondition(tt.args.condition, tt.args.stepConfig, tt.args.stepName)
			if err != nil {
				if tt.wantErr {
					if err.Error() != tt.wantErrMsg {
						t.Errorf("validateCondition() errorMsg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
						return
					}
				} else {
					t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVaidation != tt.wantVaidation {
				t.Errorf("validateCondition() = %v, want %v", gotVaidation, tt.wantVaidation)
			}
		})
	}
}

func Test_validateCondition_configExists(t *testing.T) {

	var testConfig = make(map[string]interface{})
	testConfig["emptyVal"] = ""
	testConfig["configEquals"] = "equalsTrue"
	testConfig["configExists"] = "configExistsValue"
	testConfig["filePatternKey"] = "**/conditions_test.go"
	testConfig["notStringVal"] = Config{}

	type args struct {
		condition  v1beta1.PipelineTaskCondition
		stepConfig StepConfig
		stepName   string
	}
	tests := []struct {
		name          string
		args          args
		wantVaidation bool
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "conditionRefConfigExists config invalid param count",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-exists",
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation for testStep has unsupported number of string equal parameters (1 required, got 0)").Error(),
		},
		{
			name: "conditionRefConfigExists config invalid param key",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-exists",
					Params: []v1beta1.Param{
						{Name: "invalidParam", Value: *v1beta1.NewArrayOrString("123")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation config-exists for testStep has unsupported param name: \"invalidParam\"").Error(),
		},
		{
			name: "conditionRefConfigExists config invalid empty param value",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-exists",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation config-exists for testStep has empty value").Error(),
		},
		{
			name: "conditionRefConfigExists config value does not exist",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-exists",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("configDoesNotExist")},
					},
				},
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: false,
			wantErr:       false,
		},
		{
			name: "conditionRefConfigExists config value exists",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-exists",
					Params: []v1beta1.Param{
						{Name: "configKey", Value: *v1beta1.NewArrayOrString("configExists")},
					},
				},
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: true,
			wantErr:       false,
		},
		{
			name: "conditionRefConfigExists stepName is not configured",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-exists",
					Params: []v1beta1.Param{
						{Name: "stepName", Value: *v1beta1.NewArrayOrString("testStepConfigStepName")},
					},
				},
				stepName: "testStepConfigStepName",
				stepConfig: StepConfig{
					Config: nil,
				},
			},
			wantVaidation: false,
			wantErr:       false,
		},
		{
			name: "conditionRefConfigExists stepName is configured",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "config-exists",
					Params: []v1beta1.Param{
						{Name: "stepName", Value: *v1beta1.NewArrayOrString("testStepConfigStepName")},
					},
				},
				stepName: "testStepConfigStepName",
				stepConfig: StepConfig{
					Config: map[string]interface{}{
						"steps": map[string]interface{}{
							"testStepConfigStepName": map[string]interface{}{
								"configKey": "configValue",
							},
						},
					},
				},
			},
			wantVaidation: true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVaidation, err := validateCondition(tt.args.condition, tt.args.stepConfig, tt.args.stepName)
			if err != nil {
				if tt.wantErr {
					if err.Error() != tt.wantErrMsg {
						t.Errorf("validateCondition() errorMsg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
						return
					}
				} else {
					t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVaidation != tt.wantVaidation {
				t.Errorf("validateCondition() = %v, want %v", gotVaidation, tt.wantVaidation)
			}
		})
	}
}
func Test_validateCondition_filePattern(t *testing.T) {

	var testConfig = make(map[string]interface{})
	testConfig["emptyVal"] = ""
	testConfig["configEquals"] = "equalsTrue"
	testConfig["configExists"] = "configExistsValue"
	testConfig["filePatternKey"] = "**/conditions_test.go"
	testConfig["notStringVal"] = Config{}

	type args struct {
		condition  v1beta1.PipelineTaskCondition
		stepConfig StepConfig
		stepName   string
	}
	tests := []struct {
		name          string
		args          args
		wantVaidation bool
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "conditionRefFilePattern invalid config param count",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "file-exists",
					Params: []v1beta1.Param{
						{Name: "one", Value: *v1beta1.NewArrayOrString("")},
						{Name: "two", Value: *v1beta1.NewArrayOrString("")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation for testStep has unsupported number of string equal parameters (1 required, got 2)").Error(),
		},
		{
			name: "conditionRefFilePattern invalid config invalid param key",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "file-exists",
					Params: []v1beta1.Param{
						{Name: "invalidParamKey", Value: *v1beta1.NewArrayOrString("")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation file-exists for testStep has unsupported param name: \"invalidParamKey\"").Error(),
		},
		{
			name: "conditionRefFilePattern invalid config param empty",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "file-exists",
					Params: []v1beta1.Param{
						{Name: "filePattern", Value: *v1beta1.NewArrayOrString("")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation file-exists for testStep has empty value").Error(),
		},
		{
			name: "conditionRefFilePattern file not found",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "file-exists",
					Params: []v1beta1.Param{
						{Name: "filePattern", Value: *v1beta1.NewArrayOrString("file.txt")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       false,
		},
		{
			name: "conditionRefFilePattern file found",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "file-exists",
					Params: []v1beta1.Param{
						{Name: "filePattern", Value: *v1beta1.NewArrayOrString("conditions.go")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: true,
			wantErr:       false,
		},
		{
			name: "conditionRefFilePattern file pattern from config found",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "file-exists",
					Params: []v1beta1.Param{
						{Name: "filePatternFromConfig", Value: *v1beta1.NewArrayOrString("filePatternKey")},
					},
				},
				stepName: "testStep",
				stepConfig: StepConfig{
					Config: testConfig,
				},
			},
			wantVaidation: true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVaidation, err := validateCondition(tt.args.condition, tt.args.stepConfig, tt.args.stepName)
			if err != nil {
				if tt.wantErr {
					if err.Error() != tt.wantErrMsg {
						t.Errorf("validateCondition() errorMsg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
						return
					}
				} else {
					t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVaidation != tt.wantVaidation {
				t.Errorf("validateCondition() = %v, want %v", gotVaidation, tt.wantVaidation)
			}
		})
	}
}

func Test_validateCondition_isEquasl(t *testing.T) {

	var testConfig = make(map[string]interface{})
	testConfig["emptyVal"] = ""
	testConfig["configEquals"] = "equalsTrue"
	testConfig["configExists"] = "configExistsValue"
	testConfig["filePatternKey"] = "**/conditions_test.go"
	testConfig["notStringVal"] = Config{}

	type args struct {
		condition  v1beta1.PipelineTaskCondition
		stepConfig StepConfig
		stepName   string
	}
	tests := []struct {
		name          string
		args          args
		wantVaidation bool
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "conditionRefisActive invalid config invalid param count",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "is-active",
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation for testStep has unsupported number of parameters (1 required, got 0)").Error(),
		},
		{
			name: "conditionRefisActive invalid config invalid param name",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "is-active",
					Params: []v1beta1.Param{
						{Name: "invalidParamName", Value: *v1beta1.NewArrayOrString("some value")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation is-active for testStep has unsupported param name: \"invalidParamName\"").Error(),
		},
		{
			name: "conditionRefisActive invalid config empty param value",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "is-active",
					Params: []v1beta1.Param{
						{Name: "activation", Value: *v1beta1.NewArrayOrString("")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("step condition validation is-active for testStep has empty value").Error(),
		},
		{
			name: "conditionRefisActive invalid config param value non bool",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "is-active",
					Params: []v1beta1.Param{
						{Name: "activation", Value: *v1beta1.NewArrayOrString("non-boolean value")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       true,
			wantErrMsg:    errors.Errorf("validateIsActive failed for step testStep: strconv.ParseBool: parsing \"non-boolean value\": invalid syntax").Error(),
		},
		{
			name: "conditionRefisActive invalid config param value non bool",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "is-active",
					Params: []v1beta1.Param{
						{Name: "activation", Value: *v1beta1.NewArrayOrString("false")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: false,
			wantErr:       false,
		},
		{
			name: "conditionRefisActive invalid config param value non bool",
			args: args{
				condition: v1beta1.PipelineTaskCondition{
					ConditionRef: "is-active",
					Params: []v1beta1.Param{
						{Name: "activation", Value: *v1beta1.NewArrayOrString("true")},
					},
				},
				stepName: "testStep",
			},
			wantVaidation: true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVaidation, err := validateCondition(tt.args.condition, tt.args.stepConfig, tt.args.stepName)
			if err != nil {
				if tt.wantErr {
					if err.Error() != tt.wantErrMsg {
						t.Errorf("validateCondition() errorMsg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
						return
					}
				} else {
					t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVaidation != tt.wantVaidation {
				t.Errorf("validateCondition() = %v, want %v", gotVaidation, tt.wantVaidation)
			}
		})
	}
}
