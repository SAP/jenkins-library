package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type helmMockUtilsBundle struct {
	*mock.FilesMock
	*mock.ExecMockRunner
}

func newHelmMockUtilsBundle() helmMockUtilsBundle {
	utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
	return utils
}

func TestRunHelmExecute(t *testing.T) {
	t.Parallel()
	utils := newHelmMockUtilsBundle()

	testTable := []struct {
		config         helmExecuteOptions
		expectedConfig []string
		expectedError  bool
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				ChartPath:      ".",
				DeploymentName: "testPackage",
				DeployCommand:  "test",
			},
			expectedConfig: []string{"test", "."},
			expectedError:  false,
		},
		{
			config: helmExecuteOptions{
				DeploymentName: "testPackage",
				DeployCommand:  "test",
			},
			expectedConfig: []string{"test", "."},
			expectedError:  true,
			expectedErrStr: "failed to execute helm test",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			testTable[i].config.DeployTool = "helm3"
			err := runHelmExecute(testCase.config, utils, log.Writer())
			if testCase.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedErrStr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
			}

		})
	}
}
