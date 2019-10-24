package config

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
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

	filters := metadata1.GetContextParameterFilters()

	t.Run("Secrets", func(t *testing.T) {
		for _, s := range metadata1.Spec.Inputs.Secrets {
			t.Run("All", func(t *testing.T) {
				if !sliceContains(filters.All, s.Name) {
					t.Errorf("Creation of context filter failed, expected: %v to be contained", s.Name)
				}
			})
			t.Run("General", func(t *testing.T) {
				if !sliceContains(filters.General, s.Name) {
					t.Errorf("Creation of context filter failed, expected: %v to be contained", s.Name)
				}
			})
			t.Run("Step", func(t *testing.T) {
				if !sliceContains(filters.Steps, s.Name) {
					t.Errorf("Creation of context filter failed, expected: %v to be contained", s.Name)
				}
			})
			t.Run("Stages", func(t *testing.T) {
				if !sliceContains(filters.Steps, s.Name) {
					t.Errorf("Creation of context filter failed, expected: %v to be contained", s.Name)
				}
			})
			t.Run("Parameters", func(t *testing.T) {
				if !sliceContains(filters.Parameters, s.Name) {
					t.Errorf("Creation of context filter failed, expected: %v to be contained", s.Name)
				}
			})
			t.Run("Env", func(t *testing.T) {
				if !sliceContains(filters.Env, s.Name) {
					t.Errorf("Creation of context filter failed, expected: %v to be contained", s.Name)
				}
			})
		}
	})
}
