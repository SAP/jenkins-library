package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/kubernetes/mocks"
	"github.com/stretchr/testify/assert"
)

// type helmMockUtilsBundle struct {
// 	*mock.FilesMock
// 	*mock.ExecMockRunner
// }

// func newHelmMockUtilsBundle() helmMockUtilsBundle {
// 	utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
// 	return utils
// }

func TestRunHelmExecute(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		expectedConfig []string
		expectedError  error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				ChartPath:            ".",
				DeploymentName:       "testPackage",
				HelmCommand:          "test",
				AdditionalParameters: []string{},
			},
			expectedConfig: []string{"test", "."},
			expectedError:  nil,
		},
		// {
		// 	config: helmExecuteOptions{
		// 		ChartPath:            ".",
		// 		DeploymentName:       "testPackage",
		// 		HelmCommand:          "test",
		// 		AdditionalParameters: []string{},
		// 	},
		// 	expectedConfig: []string{"test", "."},
		// 	expectedError:  errors.New("some error"),
		// },
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmTest").Return(testCase.expectedError)

			err := runHelmExecute(testCase.config.HelmCommand, testCase.config.AdditionalParameters, helmExecute)
			assert.Equal(t, testCase.expectedError, err)
		})

	}
}
