package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/kubernetes/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRunHelmUpgrade(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		expectedConfig []string
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand:          "upgrade",
				AdditionalParameters: []string{},
			},
			expectedConfig: []string{"test", "."},
			methodError:    nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand:          "upgrade",
				AdditionalParameters: []string{},
			},
			expectedConfig: []string{"test", "."},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute upgrade: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmUpgrade").Return(testCase.methodError)

			err := runHelmExecute(testCase.config.HelmCommand, testCase.config.AdditionalParameters, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmTest(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		expectedConfig []string
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand:          "test",
				AdditionalParameters: []string{},
			},
			expectedConfig: []string{"test", "."},
			methodError:    nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand:          "test",
				AdditionalParameters: []string{},
			},
			expectedConfig: []string{"test", "."},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm test: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmTest").Return(testCase.methodError)

			err := runHelmExecute(testCase.config.HelmCommand, testCase.config.AdditionalParameters, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}
