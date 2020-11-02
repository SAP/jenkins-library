package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPipelineStepData(t *testing.T) {
	var s StepData

	t.Run("Success case", func(t *testing.T) {
		myMeta := strings.NewReader("metadata:\n  name: testIt\nspec:\n  inputs:\n    params:\n      - name: testParamName\n    secrets:\n      - name: testSecret")
		err := s.ReadPipelineStepData(ioutil.NopCloser(myMeta)) // NopCloser "no-ops" the closing interface since strings do not need to be closed

		if err != nil {
			t.Errorf("Got error although no error expected: %v", err)
		}

		t.Run("step name", func(t *testing.T) {
			if s.Metadata.Name != "testIt" {
				t.Errorf("Meta name - got: %v, expected: %v", s.Metadata.Name, "testIt")
			}
		})

		t.Run("param name", func(t *testing.T) {
			if s.Spec.Inputs.Parameters[0].Name != "testParamName" {
				t.Errorf("Step name - got: %v, expected: %v", s.Spec.Inputs.Parameters[0].Name, "testParamName")
			}
		})

		t.Run("secret name", func(t *testing.T) {
			if s.Spec.Inputs.Secrets[0].Name != "testSecret" {
				t.Errorf("Step name - got: %v, expected: %v", s.Spec.Inputs.Secrets[0].Name, "testSecret")
			}
		})
	})

	t.Run("Read failure", func(t *testing.T) {
		var rc errReadCloser
		err := s.ReadPipelineStepData(rc)
		if err == nil {
			t.Errorf("Got no error although error expected.")
		}
	})

	t.Run("Unmarshalling failure", func(t *testing.T) {
		myMeta := strings.NewReader("metadata:\n\tname: testIt")
		err := s.ReadPipelineStepData(ioutil.NopCloser(myMeta))
		if err == nil {
			t.Errorf("Got no error although error expected.")
		}
	})
}

func TestGetParameterFilters(t *testing.T) {
	metadata1 := StepData{
		Spec: StepSpec{
			Inputs: StepInputs{
				Parameters: []StepParameters{
					{Name: "paramOne", Scope: []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS", "ENV"}},
					{Name: "paramTwo", Scope: []string{"STEPS", "STAGES", "PARAMETERS", "ENV"}},
					{Name: "paramThree", Scope: []string{"STAGES", "PARAMETERS", "ENV"}},
					{Name: "paramFour", Scope: []string{"PARAMETERS", "ENV"}},
					{Name: "paramFive", Scope: []string{"ENV"}},
					{Name: "paramSix"},
					{Name: "paramSeven", Scope: []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"}, Conditions: []Condition{{Params: []Param{{Name: "buildTool", Value: "mta"}}}}},
				},
			},
		},
	}

	metadata2 := StepData{
		Spec: StepSpec{
			Inputs: StepInputs{
				Parameters: []StepParameters{
					{Name: "paramOne", Scope: []string{"GENERAL"}},
					{Name: "paramTwo", Scope: []string{"STEPS"}},
					{Name: "paramThree", Scope: []string{"STAGES"}},
					{Name: "paramFour", Scope: []string{"PARAMETERS"}},
					{Name: "paramFive", Scope: []string{"ENV"}},
					{Name: "paramSix"},
				},
			},
		},
	}

	metadata3 := StepData{
		Spec: StepSpec{
			Inputs: StepInputs{
				Parameters: []StepParameters{},
			},
		},
	}

	testTable := []struct {
		Metadata              StepData
		ExpectedAll           []string
		ExpectedGeneral       []string
		ExpectedStages        []string
		ExpectedSteps         []string
		ExpectedParameters    []string
		ExpectedEnv           []string
		NotExpectedAll        []string
		NotExpectedGeneral    []string
		NotExpectedStages     []string
		NotExpectedSteps      []string
		NotExpectedParameters []string
		NotExpectedEnv        []string
	}{
		{
			Metadata:              metadata1,
			ExpectedGeneral:       []string{"verbose", "paramOne", "paramSeven", "mta"},
			ExpectedSteps:         []string{"verbose", "paramOne", "paramTwo", "paramSeven", "mta"},
			ExpectedStages:        []string{"verbose", "paramOne", "paramTwo", "paramThree", "paramSeven", "mta"},
			ExpectedParameters:    []string{"verbose", "paramOne", "paramTwo", "paramThree", "paramFour", "paramSeven", "mta"},
			ExpectedEnv:           []string{"verbose", "paramOne", "paramTwo", "paramThree", "paramFour", "paramFive", "paramSeven", "mta"},
			ExpectedAll:           []string{"verbose", "paramOne", "paramTwo", "paramThree", "paramFour", "paramFive", "paramSix", "paramSeven", "mta"},
			NotExpectedGeneral:    []string{"paramTwo", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedSteps:      []string{"paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedStages:     []string{"paramFour", "paramFive", "paramSix"},
			NotExpectedParameters: []string{"paramFive", "paramSix"},
			NotExpectedEnv:        []string{"verbose", "paramSix", "mta"},
			NotExpectedAll:        []string{},
		},
		{
			Metadata:              metadata2,
			ExpectedGeneral:       []string{"verbose", "paramOne"},
			ExpectedSteps:         []string{"verbose", "paramTwo"},
			ExpectedStages:        []string{"verbose", "paramThree"},
			ExpectedParameters:    []string{"verbose", "paramFour"},
			ExpectedEnv:           []string{"paramFive"},
			ExpectedAll:           []string{"verbose", "paramOne", "paramTwo", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedGeneral:    []string{"paramTwo", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedSteps:      []string{"paramOne", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedStages:     []string{"paramOne", "paramTwo", "paramFour", "paramFive", "paramSix"},
			NotExpectedParameters: []string{"paramOne", "paramTwo", "paramThree", "paramFive", "paramSix"},
			NotExpectedEnv:        []string{"verbose", "paramOne", "paramTwo", "paramThree", "paramFour", "paramSix"},
			NotExpectedAll:        []string{},
		},
		{
			Metadata:           metadata3,
			ExpectedGeneral:    []string{"verbose"},
			ExpectedStages:     []string{"verbose"},
			ExpectedSteps:      []string{"verbose"},
			ExpectedParameters: []string{"verbose"},
			ExpectedEnv:        []string{},
			ExpectedAll:        []string{"verbose"},
		},
	}

	for key, row := range testTable {
		t.Run(fmt.Sprintf("Metadata%v", key), func(t *testing.T) {
			filters := row.Metadata.GetParameterFilters()
			t.Run("General", func(t *testing.T) {
				for _, val := range filters.General {
					if !sliceContains(row.ExpectedGeneral, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v to be contained in  %v", val, filters.General)
					}
					if sliceContains(row.NotExpectedGeneral, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v NOT to be contained in  %v", val, filters.General)
					}
				}
			})
			t.Run("Steps", func(t *testing.T) {
				for _, val := range filters.Steps {
					if !sliceContains(row.ExpectedSteps, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v to be contained in  %v", val, filters.Steps)
					}
					if sliceContains(row.NotExpectedSteps, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v NOT to be contained in  %v", val, filters.Steps)
					}
				}
			})
			t.Run("Stages", func(t *testing.T) {
				for _, val := range filters.Stages {
					if !sliceContains(row.ExpectedStages, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v to be contained in  %v", val, filters.Stages)
					}
					if sliceContains(row.NotExpectedStages, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v NOT to be contained in  %v", val, filters.Stages)
					}
				}
			})
			t.Run("Parameters", func(t *testing.T) {
				for _, val := range filters.Parameters {
					if !sliceContains(row.ExpectedParameters, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v to be contained in  %v", val, filters.Parameters)
					}
					if sliceContains(row.NotExpectedParameters, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v NOT to be contained in  %v", val, filters.Parameters)
					}
				}
			})
			t.Run("Env", func(t *testing.T) {
				for _, val := range filters.Env {
					if !sliceContains(row.ExpectedEnv, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v to be contained in  %v", val, filters.Env)
					}
					if sliceContains(row.NotExpectedEnv, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v NOT to be contained in  %v", val, filters.Env)
					}
				}
			})
			t.Run("All", func(t *testing.T) {
				for _, val := range filters.All {
					if !sliceContains(row.ExpectedAll, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v to be contained in  %v", val, filters.All)
					}
					if sliceContains(row.NotExpectedAll, val) {
						t.Errorf("Creation of parameter filter failed, expected: %v NOT to be contained in  %v", val, filters.All)
					}
				}
			})
		})
	}
}

func TestGetContextParameterFilters(t *testing.T) {
	metadata1 := StepData{
		Spec: StepSpec{
			Inputs: StepInputs{
				Secrets: []StepSecrets{
					{Name: "testSecret1", Type: "jenkins"},
					{Name: "testSecret2", Type: "jenkins"},
				},
				Resources: []StepResources{
					{Name: "buildDescriptor", Type: "stash"},
				},
			},
		},
	}

	metadata2 := StepData{
		Spec: StepSpec{
			Containers: []Container{
				{Name: "testcontainer"},
				{Conditions: []Condition{
					{Params: []Param{
						{Name: "scanType", Value: "pip"},
					}},
				}},
			},
		},
	}

	metadata3 := StepData{
		Spec: StepSpec{
			Sidecars: []Container{
				{Name: "testsidecar"},
			},
		},
	}

	metadata4 := StepData{
		Spec: StepSpec{
			Inputs: StepInputs{
				Parameters: []StepParameters{
					{ResourceRef: []ResourceReference{{Type: "vaultSecret"}}},
				},
			},
		},
	}

	t.Run("Secrets and stashes", func(t *testing.T) {
		filters := metadata1.GetContextParameterFilters()
		assert.Equal(t, []string{"testSecret1", "testSecret2", "stashContent"}, filters.All, "incorrect filter All")
		assert.Equal(t, []string{"testSecret1", "testSecret2", "stashContent"}, filters.General, "incorrect filter General")
		assert.Equal(t, []string{"testSecret1", "testSecret2", "stashContent"}, filters.Steps, "incorrect filter Steps")
		assert.Equal(t, []string{"testSecret1", "testSecret2", "stashContent"}, filters.Stages, "incorrect filter Stages")
		assert.Equal(t, []string{"testSecret1", "testSecret2", "stashContent"}, filters.Parameters, "incorrect filter Parameters")
		assert.Equal(t, []string{"testSecret1", "testSecret2", "stashContent"}, filters.Env, "incorrect filter Env")
	})

	t.Run("Containers", func(t *testing.T) {
		filters := metadata2.GetContextParameterFilters()
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerName", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace", "pip", "scanType"}, filters.All, "incorrect filter All")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerName", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace", "pip", "scanType"}, filters.General, "incorrect filter General")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerName", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace", "pip", "scanType"}, filters.Steps, "incorrect filter Steps")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerName", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace", "pip", "scanType"}, filters.Stages, "incorrect filter Stages")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerName", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace", "pip", "scanType"}, filters.Parameters, "incorrect filter Parameters")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerName", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace", "pip", "scanType"}, filters.Env, "incorrect filter Env")
	})

	t.Run("Sidecars", func(t *testing.T) {
		filters := metadata3.GetContextParameterFilters()
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.All, "incorrect filter All")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.General, "incorrect filter General")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Steps, "incorrect filter Steps")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Stages, "incorrect filter Stages")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Parameters, "incorrect filter Parameters")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Env, "incorrect filter Env")
	})

	t.Run("Vault", func(t *testing.T) {
		filters := metadata4.GetContextParameterFilters()
		assert.Equal(t, []string{"vaultAppRoleTokenCredentialsId", "vaultAppRoleSecretTokenCredentialsId"}, filters.All, "incorrect filter All")
		assert.Equal(t, []string{"vaultAppRoleTokenCredentialsId", "vaultAppRoleSecretTokenCredentialsId"}, filters.General, "incorrect filter General")
		assert.Equal(t, []string{"vaultAppRoleTokenCredentialsId", "vaultAppRoleSecretTokenCredentialsId"}, filters.Steps, "incorrect filter Steps")
		assert.Equal(t, []string{"vaultAppRoleTokenCredentialsId", "vaultAppRoleSecretTokenCredentialsId"}, filters.Stages, "incorrect filter Stages")
		assert.Equal(t, []string{"vaultAppRoleTokenCredentialsId", "vaultAppRoleSecretTokenCredentialsId"}, filters.Parameters, "incorrect filter Parameters")
		assert.Equal(t, []string{"vaultAppRoleTokenCredentialsId", "vaultAppRoleSecretTokenCredentialsId"}, filters.Env, "incorrect filter Env")
	})
}

func TestGetContextDefaults(t *testing.T) {

	t.Run("Positive case", func(t *testing.T) {
		metadata := StepData{
			Spec: StepSpec{
				Inputs: StepInputs{
					Resources: []StepResources{
						{
							Name: "buildDescriptor",
							Type: "stash",
							Conditions: []Condition{
								{Params: []Param{
									{Name: "scanType", Value: "abc"},
								}},
							},
						},
						{
							Name: "source",
							Type: "stash",
							Conditions: []Condition{
								{Params: []Param{
									{Name: "scanType", Value: "abc"},
								}},
							},
						},
						{
							Name: "test",
							Type: "nonce",
						},
						{
							Name: "test2",
							Type: "stash",
							Conditions: []Condition{
								{Params: []Param{
									{Name: "scanType", Value: "def"},
								}},
							},
						},
						{
							Name: "test3",
							Type: "stash",
						},
					},
				},
				Containers: []Container{
					{
						Command: []string{"test/command"},
						EnvVars: []EnvVar{
							{Name: "env1", Value: "val1"},
							{Name: "env2", Value: "val2"},
						},
						Name:       "testcontainer",
						Image:      "testImage:tag",
						Shell:      "/bin/bash",
						WorkingDir: "/test/dir",
						Options: []Option{
							{Name: "opt1", Value: "optValue1"},
							{Name: "opt2", Value: "optValue2"},
						},
						//VolumeMounts: []VolumeMount{
						//	{MountPath: "mp1", Name: "mn1"},
						//	{MountPath: "mp2", Name: "mn2"},
						//},
					},
				},
				Sidecars: []Container{
					{
						Command: []string{"/sidecar/command"},
						EnvVars: []EnvVar{
							{Name: "env3", Value: "val3"},
							{Name: "env4", Value: "val4"},
						},
						Name:            "testsidecar",
						Image:           "testSidecarImage:tag",
						ImagePullPolicy: "Never",
						ReadyCommand:    "/sidecar/command",
						WorkingDir:      "/sidecar/dir",
						Options: []Option{
							{Name: "opt3", Value: "optValue3"},
							{Name: "opt4", Value: "optValue4"},
						},
						//VolumeMounts: []VolumeMount{
						//	{MountPath: "mp3", Name: "mn3"},
						//	{MountPath: "mp4", Name: "mn4"},
						//},
					},
				},
			},
		}

		cd, err := metadata.GetContextDefaults("testStep")

		t.Run("No error", func(t *testing.T) {
			if err != nil {
				t.Errorf("No error expected but got error '%v'", err)
			}
		})

		var d PipelineDefaults
		d.ReadPipelineDefaults([]io.ReadCloser{cd})

		assert.Equal(t, []interface{}{"buildDescriptor", "source"}, d.Defaults[0].Steps["testStep"]["abc"].(map[string]interface{})["stashContent"], "stashContent default not available")
		assert.Equal(t, []interface{}{"test2"}, d.Defaults[0].Steps["testStep"]["def"].(map[string]interface{})["stashContent"], "stashContent default not available")
		assert.Equal(t, []interface{}{"test3"}, d.Defaults[0].Steps["testStep"]["stashContent"], "stashContent default not available")
		assert.Equal(t, "test/command", d.Defaults[0].Steps["testStep"]["containerCommand"], "containerCommand default not available")
		assert.Equal(t, "testcontainer", d.Defaults[0].Steps["testStep"]["containerName"], "containerName default not available")
		assert.Equal(t, "/bin/bash", d.Defaults[0].Steps["testStep"]["containerShell"], "containerShell default not available")
		assert.Equal(t, map[string]interface{}{"env1": "val1", "env2": "val2"}, d.Defaults[0].Steps["testStep"]["dockerEnvVars"], "dockerEnvVars default not available")
		assert.Equal(t, "testImage:tag", d.Defaults[0].Steps["testStep"]["dockerImage"], "dockerImage default not available")
		assert.Equal(t, "testcontainer", d.Defaults[0].Steps["testStep"]["dockerName"], "dockerName default not available")
		assert.Equal(t, true, d.Defaults[0].Steps["testStep"]["dockerPullImage"], "dockerPullImage default not available")
		assert.Equal(t, "/test/dir", d.Defaults[0].Steps["testStep"]["dockerWorkspace"], "dockerWorkspace default not available")
		assert.Equal(t, []interface{}{"opt1 optValue1", "opt2 optValue2"}, d.Defaults[0].Steps["testStep"]["dockerOptions"], "dockerOptions default not available")
		//assert.Equal(t, []interface{}{"mn1:mp1", "mn2:mp2"}, d.Defaults[0].Steps["testStep"]["dockerVolumeBind"], "dockerVolumeBind default not available")

		assert.Equal(t, "/sidecar/command", d.Defaults[0].Steps["testStep"]["sidecarCommand"], "sidecarCommand default not available")
		assert.Equal(t, map[string]interface{}{"env3": "val3", "env4": "val4"}, d.Defaults[0].Steps["testStep"]["sidecarEnvVars"], "sidecarEnvVars default not available")
		assert.Equal(t, "testSidecarImage:tag", d.Defaults[0].Steps["testStep"]["sidecarImage"], "sidecarImage default not available")
		assert.Equal(t, "testsidecar", d.Defaults[0].Steps["testStep"]["sidecarName"], "sidecarName default not available")
		assert.Equal(t, false, d.Defaults[0].Steps["testStep"]["sidecarPullImage"], "sidecarPullImage default not available")
		assert.Equal(t, "/sidecar/command", d.Defaults[0].Steps["testStep"]["sidecarReadyCommand"], "sidecarReadyCommand default not available")
		assert.Equal(t, "/sidecar/dir", d.Defaults[0].Steps["testStep"]["sidecarWorkspace"], "sidecarWorkspace default not available")
		assert.Equal(t, []interface{}{"opt3 optValue3", "opt4 optValue4"}, d.Defaults[0].Steps["testStep"]["sidecarOptions"], "sidecarOptions default not available")
		//assert.Equal(t, []interface{}{"mn3:mp3", "mn4:mp4"}, d.Defaults[0].Steps["testStep"]["sidecarVolumeBind"], "sidecarVolumeBind default not available")
	})

	t.Run("Container conditions", func(t *testing.T) {
		metadata := StepData{
			Spec: StepSpec{
				Inputs: StepInputs{
					Parameters: []StepParameters{
						{Name: "testParameter", Default: "test"},
						{Name: "testConditionParameter", Default: "testConditionMet"},
					},
				},
				Containers: []Container{
					{
						Image: "testImage1:tag",
						Conditions: []Condition{
							{
								ConditionRef: "strings-equal",
								Params:       []Param{{Name: "testConditionParameter", Value: "testConditionNotMet"}},
							},
						},
					},
					{
						Image: "testImage2:tag",
						Conditions: []Condition{
							{
								ConditionRef: "strings-equal",
								Params:       []Param{{Name: "testConditionParameter", Value: "testConditionMet"}},
							},
						},
					},
				},
			},
		}

		cd, err := metadata.GetContextDefaults("testStep")

		assert.NoError(t, err)

		var d PipelineDefaults
		d.ReadPipelineDefaults([]io.ReadCloser{cd})

		assert.Equal(t, "testConditionMet", d.Defaults[0].Steps["testStep"]["testConditionParameter"])
		assert.Nil(t, d.Defaults[0].Steps["testStep"]["dockerImage"])

		metParameter := d.Defaults[0].Steps["testStep"]["testConditionMet"].(map[string]interface{})
		assert.Equal(t, "testImage2:tag", metParameter["dockerImage"])

		notMetParameter := d.Defaults[0].Steps["testStep"]["testConditionNotMet"].(map[string]interface{})
		assert.Equal(t, "testImage1:tag", notMetParameter["dockerImage"])
	})

	t.Run("Negative case", func(t *testing.T) {
		metadataErr := []StepData{
			{},
			{
				Spec: StepSpec{},
			},
			{
				Spec: StepSpec{
					Containers: []Container{},
					Sidecars:   []Container{},
				},
			},
		}

		t.Run("No containers/sidecars", func(t *testing.T) {
			cd, _ := metadataErr[0].GetContextDefaults("testStep")

			var d PipelineDefaults
			d.ReadPipelineDefaults([]io.ReadCloser{cd})

			//no assert since we just want to make sure that no panic occurs
		})

		t.Run("No command", func(t *testing.T) {
			cd, _ := metadataErr[1].GetContextDefaults("testStep")

			var d PipelineDefaults
			d.ReadPipelineDefaults([]io.ReadCloser{cd})

			//no assert since we just want to make sure that no panic occurs
		})
	})
}

func TestGetResourceParameters(t *testing.T) {
	tt := []struct {
		in       StepData
		expected map[string]interface{}
	}{
		{
			in:       StepData{Spec: StepSpec{Inputs: StepInputs{}}},
			expected: map[string]interface{}{},
		},
		{
			in: StepData{
				Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{
					{Name: "param1"},
					{Name: "param2"},
				}}}},
			expected: map[string]interface{}{},
		},
		{
			in: StepData{
				Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{
					{Name: "param1", ResourceRef: []ResourceReference{}},
					{Name: "param2", ResourceRef: []ResourceReference{}},
				}}}},
			expected: map[string]interface{}{},
		},
		{
			in: StepData{
				Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{
					{Name: "param1", ResourceRef: []ResourceReference{{Name: "notAvailable", Param: "envparam1"}}},
					{Name: "param2", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "envparam2"}}, Type: "string"},
				}}}},
			expected: map[string]interface{}{"param2": "val2"},
		},
		{
			in: StepData{
				Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{
					{Name: "param2", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "envparam2"}}, Type: "string"},
					{Name: "param3", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "jsonList"}}, Type: "[]string"},
				}}}},
			expected: map[string]interface{}{"param2": "val2", "param3": []interface{}{"value1", "value2"}},
		},
		{
			in: StepData{
				Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{
					{Name: "param4", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "jsonKeyValue"}}, Type: "map[string]interface{}"},
				}}}},
			expected: map[string]interface{}{"param4": map[string]interface{}{"key": "value"}},
		},
		{
			in: StepData{
				Spec: StepSpec{Inputs: StepInputs{Parameters: []StepParameters{
					{Name: "param1", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "envparam1"}}, Type: "noString"},
					{Name: "param4", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "jsonKeyValue"}}, Type: "string"},
				}}}},
			expected: map[string]interface{}{"param1": interface{}(nil), "param4": "{\"key\":\"value\"}"},
		},
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(dir)

	cpeDir := filepath.Join(dir, "commonPipelineEnvironment")
	err = os.MkdirAll(cpeDir, 0700)
	if err != nil {
		t.Fatal("Failed to create sub directory")
	}

	ioutil.WriteFile(filepath.Join(cpeDir, "envparam1"), []byte("val1"), 0700)
	ioutil.WriteFile(filepath.Join(cpeDir, "envparam2"), []byte("val2"), 0700)
	ioutil.WriteFile(filepath.Join(cpeDir, "jsonList"), []byte("[\"value1\",\"value2\"]"), 0700)
	ioutil.WriteFile(filepath.Join(cpeDir, "jsonKeyValue"), []byte("{\"key\":\"value\"}"), 0700)

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			got := test.in.GetResourceParameters(dir, "commonPipelineEnvironment")
			assert.Equal(t, test.expected, got)
		})
	}
}
