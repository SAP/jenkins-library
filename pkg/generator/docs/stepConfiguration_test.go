package docs

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGenerateStepDocumentationSuccess(t *testing.T) {
	// init
	testData := config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{Name: "param0", Scope: []string{"GENERAL"}, Type: "string", Default: "default0"},
					{Name: "param1", Scope: []string{"GENERAL", "STEPS"}, Type: "string", Default: "default1"},
					{Name: "param2", Scope: []string{"PARAMETERS", "STAGES", "STEPS"}, Type: "string", Default: "default2"},
				},
			},
		},
	}
	expected := "Step Configuration\n\n" +
		"We recommend to define values of step parameters via [config.yml file](../configuration.md).\n\n" +
		"In following sections of the config.yml the configuration is possible:\n\n" +
		"| parameter | general | step/stage |\n" +
		"| --------- | ------- | ---------- |\n" +
		"| `param0` | X |  |\n" +
		"| `param1` | X | X |\n" +
		"| `param2` |  | X |\n"

	// test
	result := BuildStepConfiguration(testData)

	t.Run("default", func(t *testing.T) {
		// assert
		assert.Equal(t, expected, result)
	})
	t.Run("display global parameters", func(t *testing.T) {
		t.Skip("Not yet implemented.")
		// assert
		assert.Contains(t, result, "| `noTelemetry` | X | X |\n")
		assert.Contains(t, result, "| `script` | X | X |\n")
		assert.Contains(t, result, "| `verbose` | X | X |\n")
	})
}
