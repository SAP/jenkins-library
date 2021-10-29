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
	}

	for i, test := range tt {
		t.Run(fmt.Sprintf("That arguemtns are correct %d", i), func(t *testing.T) {
			t.Parallel()
			// init
			config := test.terraformExecuteOptions
			utils := newTerraformExecuteTestsUtils()
			runner := utils.ExecMockRunner

			// test
			err := runTerraformExecute(&config, nil, utils)

			// assert
			assert.NoError(t, err)

			if config.Init {
				assert.Equal(t, mock.ExecCall{Exec: "terraform", Params: append(config.GlobalOptions, "init")}, utils.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "terraform", Params: test.expectedArgs}, utils.Calls[1])
			} else {
				assert.Equal(t, mock.ExecCall{Exec: "terraform", Params: test.expectedArgs}, utils.Calls[0])
			}

			assert.Subset(t, runner.Env, test.expectedEnvVars)
		})
	}
}
