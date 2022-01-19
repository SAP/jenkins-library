package kubernetes

import (
	"errors"
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

func TestRunHelm(t *testing.T) {

	t.Run("Helm package command", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					ChartPath:      ".",
					DeploymentName: "testPackage",
				},
				expectedConfig: []string{"package", "."},
			},
			{
				config: HelmExecuteOptions{
					ChartPath:        ".",
					DeploymentName:   "testPackage",
					PackageVersion:   "1.2.3",
					DependencyUpdate: true,
					AppVersion:       "9.8.7",
				},
				expectedConfig: []string{"package", ".", "--version", "1.2.3", "--dependency-update", "--app-version", "9.8.7"},
			},
		}

		for i, testCase := range testTable {
			err := RunHelmPackage(testCase.config, utils, log.Writer())
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm install command", func(t *testing.T) {
		t.Parallel()
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					ChartPath:             ".",
					DeploymentName:        "testPackage",
					Namespace:             "test-namespace",
					HelmDeployWaitSeconds: 525,
				},
				expectedConfig: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s"},
			},
			{
				config: HelmExecuteOptions{
					ChartPath:             ".",
					DeploymentName:        "testPackage",
					Namespace:             "test-namespace",
					HelmDeployWaitSeconds: 525,
					KeepFailedDeployments: false,
					DryRun:                true,
					AdditionalParameters:  []string{"--set-file my_script=dothings.sh"},
				},
				expectedConfig: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--dry-run", "--wait", "--timeout", "525s", "--set-file my_script=dothings.sh"},
			},
		}

		for i, testCase := range testTable {
			err := RunHelmInstall(testCase.config, utils, log.Writer())
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm uninstal command", func(t *testing.T) {
		t.Parallel()
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					ChartPath:      ".",
					DeploymentName: "testPackage",
					Namespace:      "test-namespace",
				},
				expectedConfig: []string{"uninstall", "testPackage", "--namespace", "test-namespace"},
			},
			{
				config: HelmExecuteOptions{
					ChartPath:             ".",
					DeploymentName:        "testPackage",
					Namespace:             "test-namespace",
					HelmDeployWaitSeconds: 524,
					DryRun:                true,
				},
				expectedConfig: []string{"uninstall", "testPackage", "--namespace", "test-namespace", "--wait", "--timeout", "524s", "--dry-run"},
			},
		}

		for i, testCase := range testTable {
			err := RunHelmUninstall(testCase.config, utils, log.Writer())
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm test command", func(t *testing.T) {
		t.Parallel()
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					ChartPath:      ".",
					DeploymentName: "testPackage",
				},
				expectedConfig: []string{"test", "."},
			},
			{
				config: HelmExecuteOptions{
					ChartPath:      ".",
					DeploymentName: "testPackage",
					FilterTest:     "name=test1,name=test2",
					DumpLogs:       true,
				},
				expectedConfig: []string{"test", ".", "--filter", "name=test1,name=test2", "--logs"},
			},
		}

		for i, testCase := range testTable {
			err := RunHelmTest(testCase.config, utils, log.Writer())
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm unistall command(error processing)", func(t *testing.T) {
		t.Parallel()
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config        HelmExecuteOptions
			expectedError error
		}{
			{
				config: HelmExecuteOptions{
					ChartPath:      ".",
					DeploymentName: "testPackage",
				},
				expectedError: errors.New("namespace has not been set, please configure namespace parameter"),
			},
		}

		for _, testCase := range testTable {
			err := RunHelmUninstall(testCase.config, utils, log.Writer())
			if testCase.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedError, err)
			}
		}
	})

	t.Run("Helm init", func(t *testing.T) {
		t.Parallel()
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config        HelmExecuteOptions
			expectedError error
		}{
			{
				config: HelmExecuteOptions{
					DeploymentName: "testPackage",
				},
				expectedError: errors.New("chart path has not been set, please configure chartPath parameter"),
			},
			{
				config: HelmExecuteOptions{
					ChartPath: ".",
				},
				expectedError: errors.New("deployment name has not been set, please configure deploymentName parameter"),
			},
			{
				config: HelmExecuteOptions{
					ChartPath:      ".",
					Namespace:      "test-namespace",
					DeploymentName: "testPackage",
					KubeContext:    "kubeContext",
					KubeConfig:     "kubeConfig",
				},
				expectedError: nil,
			},
		}

		for _, testCase := range testTable {
			err := runHelmInit(testCase.config, utils, log.Writer())
			if testCase.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

		}
	})

}
