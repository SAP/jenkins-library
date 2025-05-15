//go:build unit
// +build unit

package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

var stageConditionsExample string = `#Piper general purpose pipeline stage configuration including conditions
apiVersion: project-piper.io/v1
kind: PipelineDefinition
metadata:
  name: sap-piper.general.purpose.pipeline
  displayName: Piper general purpose pipeline
  description: |-
    This is a multiline
    test description
spec:
  stages:
# Init stage
  - name: init
    displayName: Init
    description: |-
      Test description
    steps:
    - name: getConfig
      description: Read pipeline stage configuration.`

var stageConditionsExpected string = `"apiVersion: project-piper.io/v1\nkind: PipelineDefinition\nmetadata:\n  description: |-\n    This is a multiline\n    test description\n  displayName: Piper general purpose pipeline\n  name: sap-piper.general.purpose.pipeline\nspec:\n` +
	`  stages:\n  - description: Test description\n    displayName: Init\n    name: init\n    steps:\n    - description: Read pipeline stage configuration.\n      name: getConfig\n"`

func defaultsOpenFileMock(name string, tokens map[string]string) (io.ReadCloser, error) {
	var r string
	switch name {
	case "TestAddCustomDefaults_default1":
		r = "default1"
	case "TestAddCustomDefaults_default2":
		r = "default3"
	case "stage_conditions.yaml":
		r = stageConditionsExample
	default:
		r = ""
	}
	return io.NopCloser(strings.NewReader(r)), nil
}

func TestDefaultsCommand(t *testing.T) {
	cmd := DefaultsCommand()

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
		exp := []string{"defaultsFile"}
		assert.Equal(t, exp, gotReq, "required flags incorrect")
	})

	t.Run("Optional flags", func(t *testing.T) {
		exp := []string{"output", "outputFile", "useV1"}
		assert.Equal(t, exp, gotOpt, "optional flags incorrect")
	})

	t.Run("Run", func(t *testing.T) {
		t.Run("Success case", func(t *testing.T) {
			defaultsOptions.openFile = defaultsOpenFileMock
			defaultsOptions.defaultsFiles = []string{"test", "test"}
			cmd.Run(cmd, []string{})
		})
	})
}

func TestGenerateDefaults(t *testing.T) {
	testParams := []struct {
		name          string
		defaultsFiles []string
		useV1         bool
		expected      string
	}{
		{
			name:          "Single defaults file",
			defaultsFiles: []string{"test"},
			expected:      `{"content":"general: null\nstages: null\nsteps: null\n","filepath":"test"}`,
		},
		{
			name:          "Multiple defaults files",
			defaultsFiles: []string{"test1", "test2"},
			expected: `[{"content":"general: null\nstages: null\nsteps: null\n","filepath":"test1"},` +
				`{"content":"general: null\nstages: null\nsteps: null\n","filepath":"test2"}]`,
		},
		{
			name:          "Single file + useV1",
			defaultsFiles: []string{"stage_conditions.yaml"},
			useV1:         true,
			expected:      `{"content":` + stageConditionsExpected + `,"filepath":"stage_conditions.yaml"}`,
		},
		{
			name:          "Multiple files + useV1",
			defaultsFiles: []string{"stage_conditions.yaml", "stage_conditions.yaml"},
			useV1:         true,
			expected: `[{"content":` + stageConditionsExpected + `,"filepath":"stage_conditions.yaml"},` +
				`{"content":` + stageConditionsExpected + `,"filepath":"stage_conditions.yaml"}]`,
		},
	}

	utils := newGetDefaultsUtilsUtils()
	defaultsOptions.openFile = defaultsOpenFileMock

	for _, test := range testParams {
		t.Run(test.name, func(t *testing.T) {
			defaultsOptions.defaultsFiles = test.defaultsFiles
			defaultsOptions.useV1 = test.useV1
			result, _ := generateDefaults(utils)
			assert.Equal(t, test.expected, string(result))
		})
	}
}
