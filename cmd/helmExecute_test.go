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
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "upgrade",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "upgrade",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute upgrade: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmUpgrade").Return(testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmLint(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		expectedConfig []string
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "lint",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "lint",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm lint: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmLint").Return(testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmInstall(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		expectedConfig []string
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "install",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "install",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm install: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmInstall").Return(testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute)
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
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "test",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "test",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm test: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmTest").Return(testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmUninstall(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "uninstall",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "uninstall",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm uninstall: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmUninstall").Return(testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmPackage(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "package",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "package",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm package: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmPackage").Return(testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmPush(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "publish",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "publish",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm publish: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmPublish").Return(testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmDefaultCommand(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config             helmExecuteOptions
		methodLintError    error
		methodPackageError error
		methodPublishError error
		expectedErrStr     string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodLintError:    nil,
			methodPackageError: nil,
			methodPublishError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodLintError: errors.New("some error"),
			expectedErrStr:  "failed to execute helm lint: some error",
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodPackageError: errors.New("some error"),
			expectedErrStr:     "failed to execute helm package: some error",
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodPublishError: errors.New("some error"),
			expectedErrStr:     "failed to execute helm publish: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmLint").Return(testCase.methodLintError)
			helmExecute.On("RunHelmPackage").Return(testCase.methodPackageError)
			helmExecute.On("RunHelmPublish").Return(testCase.methodPublishError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}

		})
	}

}
