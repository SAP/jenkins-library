package cmd

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func configOpenFileMock(name string) (io.ReadCloser, error) {
	var r string
	switch name {
	case "TestAddCustomDefaults_default1":
		r = "default1"
	case "TestAddCustomDefaults_default2":
		r = "default3"
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
}

func TestConfigCommand(t *testing.T) {
	cmd := ConfigCommand()

	gotReq := []string{}
	gotOpt := []string{}

	cmd.Flags().VisitAll(func(pflag *flag.Flag) {
		annotations, found := pflag.Annotations[cobra.BashCompOneRequiredFlag]
		if found && annotations[0] == "true" {
			gotReq = append(gotReq, pflag.Name)
		} else {
			gotOpt = append(gotOpt, pflag.Name)
		}
	})

	t.Run("Required flags", func(t *testing.T) {
		exp := []string{"stepMetadata"}
		assert.Equal(t, exp, gotReq, "required flags incorrect")
	})

	t.Run("Optional flags", func(t *testing.T) {
		exp := []string{"contextConfig", "output", "parametersJSON"}
		assert.Equal(t, exp, gotOpt, "optional flags incorrect")
	})

	t.Run("Run", func(t *testing.T) {
		t.Run("Success case", func(t *testing.T) {
			configOptions.openFile = configOpenFileMock
			err := cmd.RunE(cmd, []string{})
			assert.NoError(t, err, "error occured but none expected")
		})
	})
}

func TestDefaultsAndFilters(t *testing.T) {
	metadata := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{Name: "paramOne", Scope: []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS", "ENV"}},
				},
			},
		},
	}

	t.Run("Context config", func(t *testing.T) {
		configOptions.contextConfig = true
		defer func() { configOptions.contextConfig = false }()
		defaults, filters, err := defaultsAndFilters(&metadata)

		assert.Equal(t, 1, len(defaults), "getting defaults failed")
		assert.Equal(t, 0, len(filters.All), "wrong number of filter values")
		assert.NoError(t, err, "error occured but none expected")
	})

	t.Run("Step config", func(t *testing.T) {
		defaults, filters, err := defaultsAndFilters(&metadata)
		assert.Equal(t, 0, len(defaults), "getting defaults failed")
		assert.Equal(t, 1, len(filters.All), "wrong number of filter values")
		assert.NoError(t, err, "error occured but none expected")
	})

}
