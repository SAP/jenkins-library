//go:build unit
// +build unit

package generator

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_adjustDefaultValues(t *testing.T) {
	tests := []struct {
		want  interface{}
		name  string
		input *config.StepData
	}{
		{want: false, name: "boolean", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "bool", Mandatory: true},
		}}}}},
		{want: nil, name: "integer", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "int", Mandatory: true},
		}}}}},
		{want: nil, name: "string", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true},
		}}}}},
		{want: nil, name: "string array", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true},
		}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			adjustDefaultValues(tt.input)
			// assert
			assert.Equal(t, tt.want, tt.input.Spec.Inputs.Parameters[0].Default)
		})
	}
}

func Test_adjustMandatoryFlags(t *testing.T) {
	tests := []struct {
		want  bool
		name  string
		input *config.StepData
	}{
		{want: false, name: "boolean with empty default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "bool", Mandatory: true, Default: false},
		}}}}},
		{want: false, name: "boolean with default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "bool", Mandatory: true, Default: true},
		}}}}},
		{want: true, name: "string with default not set", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true},
		}}}}},
		{want: true, name: "string with empty default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true, Default: ""},
		}}}}},
		{want: false, name: "string with default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "string", Mandatory: true, Default: "Oktober"},
		}}}}},
		{want: true, name: "string array with default not set", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true},
		}}}}},
		{want: true, name: "string array with empty default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true, Default: []string{}},
		}}}}},
		{want: false, name: "string array with default", input: &config.StepData{Spec: config.StepSpec{Inputs: config.StepInputs{Parameters: []config.StepParameters{
			{Name: "param", Type: "[]string", Mandatory: true, Default: []string{"Oktober"}},
		}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			adjustMandatoryFlags(tt.input)
			// assert
			assert.Equal(t, tt.want, tt.input.Spec.Inputs.Parameters[0].Mandatory)
		})
	}
}

func Test_interfaceArrayLength(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "empty type",
			args: args{},
			want: -1,
		},
		{
			name: "string type",
			args: args{"string"},
			want: -1,
		},
		{
			name: "empty array type",
			args: args{[]interface{}{}},
			want: 0,
		},
		{
			name: "string array type",
			args: args{[]interface{}{"string1", "string1"}},
			want: 2,
		},
		{
			name: "string array type",
			args: args{[]string{"string1", "string1"}},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interfaceArrayLength(tt.args.i); got != tt.want {
				t.Errorf("interfaceArrayLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
