package config

import (
	"fmt"
	"io"
	"io/ioutil"
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
			ExpectedGeneral:       []string{"paramOne"},
			ExpectedSteps:         []string{"paramOne", "paramTwo"},
			ExpectedStages:        []string{"paramOne", "paramTwo", "paramThree"},
			ExpectedParameters:    []string{"paramOne", "paramTwo", "paramThree", "paramFour"},
			ExpectedEnv:           []string{"paramOne", "paramTwo", "paramThree", "paramFour", "paramFive"},
			ExpectedAll:           []string{"paramOne", "paramTwo", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedGeneral:    []string{"paramTwo", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedSteps:      []string{"paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedStages:     []string{"paramFour", "paramFive", "paramSix"},
			NotExpectedParameters: []string{"paramFive", "paramSix"},
			NotExpectedEnv:        []string{"paramSix"},
			NotExpectedAll:        []string{},
		},
		{
			Metadata:              metadata2,
			ExpectedGeneral:       []string{"paramOne"},
			ExpectedSteps:         []string{"paramTwo"},
			ExpectedStages:        []string{"paramThree"},
			ExpectedParameters:    []string{"paramFour"},
			ExpectedEnv:           []string{"paramFive"},
			ExpectedAll:           []string{"paramOne", "paramTwo", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedGeneral:    []string{"paramTwo", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedSteps:      []string{"paramOne", "paramThree", "paramFour", "paramFive", "paramSix"},
			NotExpectedStages:     []string{"paramOne", "paramTwo", "paramFour", "paramFive", "paramSix"},
			NotExpectedParameters: []string{"paramOne", "paramTwo", "paramThree", "paramFive", "paramSix"},
			NotExpectedEnv:        []string{"paramOne", "paramTwo", "paramThree", "paramFour", "paramSix"},
			NotExpectedAll:        []string{},
		},
		{
			Metadata:           metadata3,
			ExpectedGeneral:    []string{},
			ExpectedStages:     []string{},
			ExpectedSteps:      []string{},
			ExpectedParameters: []string{},
			ExpectedEnv:        []string{},
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
			},
		},
	}

	metadata2 := StepData{
		Spec: StepSpec{
			Containers: []Container{
				{Name: "testcontainer"},
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

	t.Run("Secrets", func(t *testing.T) {
		filters := metadata1.GetContextParameterFilters()
		assert.Equal(t, []string{"testSecret1", "testSecret2"}, filters.All, "incorrect filter All")
		assert.Equal(t, []string{"testSecret1", "testSecret2"}, filters.General, "incorrect filter General")
		assert.Equal(t, []string{"testSecret1", "testSecret2"}, filters.Steps, "incorrect filter Steps")
		assert.Equal(t, []string{"testSecret1", "testSecret2"}, filters.Stages, "incorrect filter Stages")
		assert.Equal(t, []string{"testSecret1", "testSecret2"}, filters.Parameters, "incorrect filter Parameters")
		assert.Equal(t, []string{"testSecret1", "testSecret2"}, filters.Env, "incorrect filter Env")
	})

	t.Run("Containers", func(t *testing.T) {
		filters := metadata2.GetContextParameterFilters()
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}, filters.All, "incorrect filter All")
		assert.NotEqual(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}, filters.General, "incorrect filter General")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}, filters.Steps, "incorrect filter Steps")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}, filters.Stages, "incorrect filter Stages")
		assert.Equal(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}, filters.Parameters, "incorrect filter Parameters")
		assert.NotEqual(t, []string{"containerCommand", "containerShell", "dockerEnvVars", "dockerImage", "dockerOptions", "dockerPullImage", "dockerVolumeBind", "dockerWorkspace"}, filters.Env, "incorrect filter Env")
	})

	t.Run("Sidecars", func(t *testing.T) {
		filters := metadata3.GetContextParameterFilters()
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.All, "incorrect filter All")
		assert.NotEqual(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.General, "incorrect filter General")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Steps, "incorrect filter Steps")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Stages, "incorrect filter Stages")
		assert.Equal(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Parameters, "incorrect filter Parameters")
		assert.NotEqual(t, []string{"containerName", "containerPortMappings", "dockerName", "sidecarEnvVars", "sidecarImage", "sidecarName", "sidecarOptions", "sidecarPullImage", "sidecarReadyCommand", "sidecarVolumeBind", "sidecarWorkspace"}, filters.Env, "incorrect filter Env")
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
						},
						{
							Name: "source",
							Type: "stash",
						},
						{
							Name: "test",
							Type: "nonce",
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

		assert.Equal(t, []interface{}{"buildDescriptor", "source"}, d.Defaults[0].Steps["testStep"]["stashContent"], "stashContent default not available")
		assert.Equal(t, "test/command", d.Defaults[0].Steps["testStep"]["containerCommand"], "containerCommand default not available")
		assert.Equal(t, "testcontainer", d.Defaults[0].Steps["testStep"]["containerName"], "containerName default not available")
		assert.Equal(t, "/bin/bash", d.Defaults[0].Steps["testStep"]["containerShell"], "containerShell default not available")
		assert.Equal(t, []interface{}{"env1=val1", "env2=val2"}, d.Defaults[0].Steps["testStep"]["dockerEnvVars"], "dockerEnvVars default not available")
		assert.Equal(t, "testImage:tag", d.Defaults[0].Steps["testStep"]["dockerImage"], "dockerImage default not available")
		assert.Equal(t, "testcontainer", d.Defaults[0].Steps["testStep"]["dockerName"], "dockerName default not available")
		assert.Equal(t, true, d.Defaults[0].Steps["testStep"]["dockerPullImage"], "dockerPullImage default not available")
		assert.Equal(t, "/test/dir", d.Defaults[0].Steps["testStep"]["dockerWorkspace"], "dockerWorkspace default not available")

		assert.Equal(t, "/sidecar/command", d.Defaults[0].Steps["testStep"]["sidecarCommand"], "sidecarCommand default not available")
		assert.Equal(t, []interface{}{"env3=val3", "env4=val4"}, d.Defaults[0].Steps["testStep"]["sidecarEnvVars"], "sidecarEnvVars default not available")
		assert.Equal(t, "testSidecarImage:tag", d.Defaults[0].Steps["testStep"]["sidecarImage"], "sidecarImage default not available")
		assert.Equal(t, "testsidecar", d.Defaults[0].Steps["testStep"]["sidecarName"], "sidecarName default not available")
		assert.Equal(t, false, d.Defaults[0].Steps["testStep"]["sidecarPullImage"], "sidecarPullImage default not available")
		assert.Equal(t, "/sidecar/command", d.Defaults[0].Steps["testStep"]["sidecarReadyCommand"], "sidecarReadyCommand default not available")
		assert.Equal(t, "/sidecar/dir", d.Defaults[0].Steps["testStep"]["sidecarWorkspace"], "sidecarWorkspace default not available")
	})

	t.Run("Negative case", func(t *testing.T) {
		metadataErr := []StepData{
			StepData{},
			StepData{
				Spec: StepSpec{},
			},
			StepData{
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
