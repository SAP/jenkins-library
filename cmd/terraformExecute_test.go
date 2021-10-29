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
		expectedArgs []string
	}{
		{
			terraformExecuteOptions{
				Command: "apply",
			}, []string{"apply", "-auto-approve"},
		},
		{
			terraformExecuteOptions{
				Command:          "apply",
				TerraformSecrets: "/tmp/test",
			}, []string{"apply", "-auto-approve", "-var-file=/tmp/test"},
		},
		{
			terraformExecuteOptions{
				Command: "plan",
			}, []string{"plan"},
		},
		{
			terraformExecuteOptions{
				Command:          "plan",
				TerraformSecrets: "/tmp/test",
			}, []string{"plan", "-var-file=/tmp/test"},
		},
		{
			terraformExecuteOptions{
				Command:          "plan",
				TerraformSecrets: "/tmp/test",
				AdditionalArgs:   []string{"-arg1"},
			}, []string{"plan", "-var-file=/tmp/test", "-arg1"},
		},
		{
			terraformExecuteOptions{
				Command:          "apply",
				TerraformSecrets: "/tmp/test",
				AdditionalArgs:   []string{"-arg1"},
			}, []string{"apply", "-auto-approve", "-var-file=/tmp/test", "-arg1"},
		},
		{
			terraformExecuteOptions{
				Command:          "apply",
				TerraformSecrets: "/tmp/test",
				AdditionalArgs:   []string{"-arg1"},
				GlobalOptions:    []string{"-chgdir=src"},
			}, []string{"-chgdir=src", "apply", "-auto-approve", "-var-file=/tmp/test", "-arg1"},
		},
	}

	for i, test := range tt {
		t.Run(fmt.Sprintf("That arguemtns are correct %d", i), func(t *testing.T) {
			t.Parallel()
			// init
			config := test.terraformExecuteOptions
			utils := newTerraformExecuteTestsUtils()

			// test
			err := runTerraformExecute(&config, nil, utils)

			// assert
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "terraform", Params: test.expectedArgs}, utils.Calls[0])
		})
	}
}
