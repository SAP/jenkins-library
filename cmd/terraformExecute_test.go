//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type terraformExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTerraformExecuteTestsUtils() terraformExecuteMockUtils {
	utils := terraformExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunTerraformExecute(t *testing.T) {
	t.Parallel()

	tt := []struct {
		terraformExecuteOptions
		expectedArgs    []string
		expectedEnvVars []string
	}{
		{
			terraformExecuteOptions{
				Command: "apply",
			}, []string{"apply", "-auto-approve", "-no-color"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command:          "apply",
				TerraformSecrets: "/tmp/test",
			}, []string{"apply", "-auto-approve", "-var-file=/tmp/test", "-no-color"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command: "plan",
			}, []string{"plan", "-no-color"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command:          "plan",
				TerraformSecrets: "/tmp/test",
			}, []string{"plan", "-var-file=/tmp/test", "-no-color"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command:          "plan",
				TerraformSecrets: "/tmp/test",
				AdditionalArgs:   []string{"-arg1"},
			}, []string{"plan", "-var-file=/tmp/test", "-no-color", "-arg1"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command:          "apply",
				TerraformSecrets: "/tmp/test",
				AdditionalArgs:   []string{"-arg1"},
			}, []string{"apply", "-auto-approve", "-var-file=/tmp/test", "-no-color", "-arg1"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command:          "apply",
				TerraformSecrets: "/tmp/test",
				AdditionalArgs:   []string{"-arg1"},
				GlobalOptions:    []string{"-chgdir=src"},
			}, []string{"-chgdir=src", "apply", "-auto-approve", "-var-file=/tmp/test", "-no-color", "-arg1"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command: "apply",
				Init:    true,
			}, []string{"apply", "-auto-approve", "-no-color"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command:       "apply",
				GlobalOptions: []string{"-chgdir=src"},
				Init:          true,
			}, []string{"-chgdir=src", "apply", "-auto-approve", "-no-color"}, []string{},
		},
		{
			terraformExecuteOptions{
				Command:       "apply",
				CliConfigFile: ".pipeline/.terraformrc",
			}, []string{"apply", "-auto-approve", "-no-color"}, []string{"TF_CLI_CONFIG_FILE=.pipeline/.terraformrc"},
		},
		{
			terraformExecuteOptions{
				Command:   "plan",
				Workspace: "any-workspace",
			}, []string{"plan", "-no-color"}, []string{"TF_WORKSPACE=any-workspace"},
		},
	}

	for i, test := range tt {
		t.Run(fmt.Sprintf("That arguments are correct %d", i), func(t *testing.T) {
			t.Parallel()
			// init
			config := test.terraformExecuteOptions
			utils := newTerraformExecuteTestsUtils()
			utils.StdoutReturn = map[string]string{}
			utils.StdoutReturn["terraform output -json"] = "{}"
			utils.StdoutReturn["terraform -chgdir=src output -json"] = "{}"

			runner := utils.ExecMockRunner

			// test
			err := runTerraformExecute(&config, nil, utils, &terraformExecuteCommonPipelineEnvironment{})

			// assert
			assert.NoError(t, err)

			if config.Init {
				assert.Equal(t, mock.ExecCall{Exec: "terraform", Params: append(config.GlobalOptions, "init", "-no-color")}, utils.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "terraform", Params: test.expectedArgs}, utils.Calls[1])
			} else {
				assert.Equal(t, mock.ExecCall{Exec: "terraform", Params: test.expectedArgs}, utils.Calls[0])
			}

			assert.Subset(t, runner.Env, test.expectedEnvVars)
		})
	}

	t.Run("Outputs get injected into CPE", func(t *testing.T) {
		t.Parallel()

		cpe := terraformExecuteCommonPipelineEnvironment{}

		config := terraformExecuteOptions{
			Command: "plan",
		}
		utils := newTerraformExecuteTestsUtils()
		utils.StdoutReturn = map[string]string{}
		utils.StdoutReturn["terraform output -json"] = `{
			"sample_var": {
				"sensitive": true,
				"value": "a secret value",
				"type": "string"
			}
}
		`

		// test
		err := runTerraformExecute(&config, nil, utils, &cpe)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, 1, len(cpe.custom.terraformOutputs))
		assert.Equal(t, "a secret value", cpe.custom.terraformOutputs["sample_var"])
	})
}
