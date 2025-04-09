//go:build unit

package kubernetes

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type helmMockUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
}

func TestRunHelmInit(t *testing.T) {
	testTable := []struct {
		config        HelmExecuteOptions
		expectedError error
	}{
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

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.runHelmInit()
			if testCase.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestRunHelmAdd(t *testing.T) {
	testTable := []struct {
		config            HelmExecuteOptions
		expectedExecCalls []mock.ExecCall
		generalVerbose    bool
		expectedError     error
	}{
		{
			config: HelmExecuteOptions{
				TargetRepositoryURL:      "https://charts.helm.sh/stable",
				TargetRepositoryName:     "stable",
				TargetRepositoryUser:     "userAccount",
				TargetRepositoryPassword: "pwdAccount",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"repo", "add", "--username", "userAccount", "--password", "pwdAccount", "stable", "https://charts.helm.sh/stable"}},
			},
			generalVerbose: false,
			expectedError:  nil,
		},
		{
			config: HelmExecuteOptions{
				TargetRepositoryURL:  "https://charts.helm.sh/stable",
				TargetRepositoryName: "test",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"repo", "add", "test", "https://charts.helm.sh/stable", "--debug"}},
			},
			generalVerbose: true,
			expectedError:  nil,
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: testCase.generalVerbose,
				stdout:  log.Writer(),
			}
			err := helmExecute.runHelmAdd(testCase.config.TargetRepositoryName, testCase.config.TargetRepositoryURL, testCase.config.TargetRepositoryUser, testCase.config.TargetRepositoryPassword)
			if testCase.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmUpgrade(t *testing.T) {
	os.Setenv("IMAGE", "image")
	os.Setenv("PIPER_VAULTCREDENTIAL_IMAGE", "image")

	testTable := []struct {
		config            HelmExecuteOptions
		generalVerbose    bool
		expectedExecCalls []mock.ExecCall
	}{
		{
			config: HelmExecuteOptions{
				DeploymentName:        "test_deployment",
				ChartPath:             "",
				Namespace:             "test_namespace",
				ForceUpdates:          true,
				HelmDeployWaitSeconds: 3456,
				AdditionalParameters:  []string{"additional", "parameters"},
				Image:                 "dtzar/helm-kubectl:3.4.1",
				TargetRepositoryName:  "test",
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
				RenderSubchartNotes:   true,
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"repo", "add", "test", "https://charts.helm.sh/stable", "--debug"}},
				{Exec: "helm", Params: []string{"upgrade", "test_deployment", "test", "--debug", "--install", "--namespace", "test_namespace", "--force", "--wait", "--timeout", "3456s", "--atomic", "--render-subchart-notes", "additional", "parameters"}},
			},
		},
		{
			config: HelmExecuteOptions{
				DeploymentName:        "test_deployment",
				ChartPath:             ".",
				Namespace:             "test_namespace",
				ForceUpdates:          true,
				HelmDeployWaitSeconds: 3456,
				AdditionalParameters:  []string{"additional", "parameters"},
				Image:                 "dtzar/helm-kubectl:3.4.1",
				TargetRepositoryName:  "test",
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"upgrade", "test_deployment", ".", "--debug", "--install", "--namespace", "test_namespace", "--force", "--wait", "--timeout", "3456s", "--atomic", "additional", "parameters"}},
			},
		},
		{
			config: HelmExecuteOptions{
				DeploymentName:        "test_deployment",
				ChartPath:             ".",
				Namespace:             "test_namespace",
				ForceUpdates:          true,
				HelmDeployWaitSeconds: 3456,
				AdditionalParameters:  []string{"--set", "image.repository=$IMAGE"},
				Image:                 "dtzar/helm-kubectl:3.4.1",
				TargetRepositoryName:  "test",
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"upgrade", "test_deployment", ".", "--debug", "--install", "--namespace", "test_namespace", "--force", "--wait", "--timeout", "3456s", "--atomic", "--set", "image.repository=$IMAGE"}},
			},
		},
		{
			config: HelmExecuteOptions{
				DeploymentName:        "test_deployment",
				ChartPath:             ".",
				Namespace:             "test_namespace",
				ForceUpdates:          true,
				HelmDeployWaitSeconds: 3456,
				AdditionalParameters:  []string{"--set", "image.repository=$PIPER_VAULTCREDENTIAL_IMAGE"},
				Image:                 "dtzar/helm-kubectl:3.4.1",
				TargetRepositoryName:  "test",
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"upgrade", "test_deployment", ".", "--debug", "--install", "--namespace", "test_namespace", "--force", "--wait", "--timeout", "3456s", "--atomic", "--set", "image.repository=image"}},
			},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: testCase.generalVerbose,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmUpgrade()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmLint(t *testing.T) {
	testTable := []struct {
		config            HelmExecuteOptions
		expectedExecCalls []mock.ExecCall
	}{
		{
			config: HelmExecuteOptions{
				ChartPath: ".",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"lint", "."}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:  ".",
				HelmValues: []string{"./values_1.yaml", "./values_2.yaml"},
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"lint", ".", "--values", "./values_1.yaml", "--values", "./values_2.yaml"}},
			},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmLint()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmInstall(t *testing.T) {
	os.Setenv("PIPER_VAULTCREDENTIAL_MY_SCRIPT", "dothings.sh")

	testTable := []struct {
		config            HelmExecuteOptions
		generalVerbose    bool
		expectedExecCalls []mock.ExecCall
	}{
		{
			config: HelmExecuteOptions{
				ChartPath:             "",
				DeploymentName:        "testPackage",
				Namespace:             "test-namespace",
				HelmDeployWaitSeconds: 525,
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
				TargetRepositoryName:  "test",
				RenderSubchartNotes:   true,
			},
			generalVerbose: false,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"repo", "add", "test", "https://charts.helm.sh/stable"}},
				{Exec: "helm", Params: []string{"install", "testPackage", "test", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s", "--render-subchart-notes"}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:             ".",
				DeploymentName:        "testPackage",
				Namespace:             "test-namespace",
				HelmDeployWaitSeconds: 525,
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
				TargetRepositoryName:  "test",
			},
			generalVerbose: false,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s"}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:             ".",
				DeploymentName:        "testPackage",
				Namespace:             "test-namespace",
				HelmDeployWaitSeconds: 525,
				KeepFailedDeployments: false,
				AdditionalParameters:  []string{"--set-file", "my_script=dothings.sh"},
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
				TargetRepositoryName:  "test",
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s", "--set-file", "my_script=dothings.sh", "--debug", "--dry-run", "--hide-secret"}},
				{Exec: "helm", Params: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s", "--set-file", "my_script=dothings.sh", "--debug"}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:             ".",
				DeploymentName:        "testPackage",
				Namespace:             "test-namespace",
				HelmDeployWaitSeconds: 525,
				KeepFailedDeployments: false,
				AdditionalParameters:  []string{"--set-file", "my_script=$PIPER_VAULTCREDENTIAL_MY_SCRIPT"},
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
				TargetRepositoryName:  "test",
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s", "--set-file", "my_script=dothings.sh", "--debug", "--dry-run", "--hide-secret"}},
				{Exec: "helm", Params: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s", "--set-file", "my_script=dothings.sh", "--debug"}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:             ".",
				DeploymentName:        "testPackage",
				Namespace:             "test-namespace",
				HelmDeployWaitSeconds: 525,
				KeepFailedDeployments: false,
				AdditionalParameters:  []string{"--set", "auth=Basic user:password"},
				TargetRepositoryURL:   "https://charts.helm.sh/stable",
				TargetRepositoryName:  "test",
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s", "--set", "auth=Basic user:password", "--debug", "--dry-run", "--hide-secret"}},
				{Exec: "helm", Params: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s", "--set", "auth=Basic user:password", "--debug"}},
			},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: testCase.generalVerbose,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmInstall()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmUninstall(t *testing.T) {
	testTable := []struct {
		config            HelmExecuteOptions
		generalVerbose    bool
		expectedExecCalls []mock.ExecCall
		expectedError     error
	}{
		{
			config: HelmExecuteOptions{
				ChartPath:            ".",
				DeploymentName:       "testPackage",
				Namespace:            "test-namespace",
				TargetRepositoryName: "test",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"uninstall", "testPackage", "--namespace", "test-namespace"}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:             ".",
				DeploymentName:        "testPackage",
				Namespace:             "test-namespace",
				HelmDeployWaitSeconds: 524,
				TargetRepositoryName:  "test",
			},
			generalVerbose: true,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"uninstall", "testPackage", "--namespace", "test-namespace", "--wait", "--timeout", "524s", "--debug", "--dry-run", "--hide-secret"}},
				{Exec: "helm", Params: []string{"uninstall", "testPackage", "--namespace", "test-namespace", "--wait", "--timeout", "524s", "--debug"}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:            ".",
				DeploymentName:       "testPackage",
				TargetRepositoryName: "test",
			},
			expectedError: errors.New("namespace has not been set, please configure namespace parameter"),
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: testCase.generalVerbose,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmUninstall()
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmPackage(t *testing.T) {
	testTable := []struct {
		config            HelmExecuteOptions
		expectedExecCalls []mock.ExecCall
	}{
		{
			config: HelmExecuteOptions{
				ChartPath:      ".",
				DeploymentName: "testPackage",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"package", "."}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:               ".",
				DeploymentName:          "testPackage",
				Version:                 "1.2.3",
				PackageDependencyUpdate: true,
				AppVersion:              "9.8.7",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"package", ".", "--version", "1.2.3", "--dependency-update", "--app-version", "9.8.7"}},
			},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.runHelmPackage()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmTest(t *testing.T) {
	testTable := []struct {
		config            HelmExecuteOptions
		expectedExecCalls []mock.ExecCall
	}{
		{
			config: HelmExecuteOptions{
				ChartPath:      ".",
				DeploymentName: "testPackage",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"test", "."}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:      ".",
				DeploymentName: "testPackage",
				FilterTest:     "name=test1,name=test2",
				DumpLogs:       true,
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"test", ".", "--filter", "name=test1,name=test2", "--logs"}},
			},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmTest()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmDependency(t *testing.T) {
	testTable := []struct {
		config            HelmExecuteOptions
		expectedError     error
		expectedExecCalls []mock.ExecCall
	}{
		{
			config: HelmExecuteOptions{
				ChartPath: ".",
			},
			expectedError:     errors.New("there is no dependency value. Possible values are build, list, update"),
			expectedExecCalls: nil,
		},
		{
			config: HelmExecuteOptions{
				ChartPath:  ".",
				Dependency: "update",
			},
			expectedError: nil,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"dependency", "update", "."}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:            ".",
				Dependency:           "update",
				SourceRepositoryName: "foo",
				SourceRepositoryURL:  "bar",
			},
			expectedError: nil,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"repo", "add", "foo", "bar"}},
				{Exec: "helm", Params: []string{"dependency", "update", "."}},
			},
		},
		{
			config: HelmExecuteOptions{
				ChartPath:                ".",
				Dependency:               "update",
				SourceRepositoryName:     "foo",
				SourceRepositoryURL:      "bar",
				SourceRepositoryUser:     "username",
				SourceRepositoryPassword: "password",
			},
			expectedError: nil,
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"repo", "add", "--username", "username", "--password", "password", "foo", "bar"}},
				{Exec: "helm", Params: []string{"dependency", "update", "."}},
			},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
				FilesMock: &mock.FilesMock{
					Separator: "/",
				},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmDependency()
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}

func TestRunHelmPublish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		utils := helmMockUtilsBundle{
			ExecMockRunner: &mock.ExecMockRunner{},
			HttpClientMock: &mock.HttpClientMock{
				FileUploads: map[string]string{},
			},
		}

		config := HelmExecuteOptions{
			TargetRepositoryURL:      "https://my.target.repository.local/",
			TargetRepositoryUser:     "testUser",
			TargetRepositoryPassword: "testPWD",
			PublishVersion:           "1.2.3",
			DeploymentName:           "test_helm_chart",
			ChartPath:                ".",
		}
		utils.ReturnFileUploadStatus = 200

		helmExecute := HelmExecute{
			utils:   utils,
			config:  config,
			verbose: false,
			stdout:  log.Writer(),
		}

		targetURL, err := helmExecute.RunHelmPublish()
		if assert.NoError(t, err) {
			assert.Equal(t, 1, len(utils.FileUploads))
			assert.Equal(t, "https://my.target.repository.local/test_helm_chart-1.2.3.tgz", targetURL)
			assert.Equal(t, "https://my.target.repository.local/test_helm_chart-1.2.3.tgz", utils.FileUploads["test_helm_chart-1.2.3.tgz"])
		}
	})
}

func TestRunHelmCommand(t *testing.T) {
	testTable := []struct {
		helmParams        []string
		config            HelmExecuteOptions
		expectedExecCalls []mock.ExecCall
	}{
		{
			helmParams: []string{"lint, package, publish"},
			config: HelmExecuteOptions{
				HelmCommand: "lint_package_publish",
			},
			expectedExecCalls: []mock.ExecCall{
				{Exec: "helm", Params: []string{"lint, package, publish"}},
			},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprintf("test case: %d", i), func(t *testing.T) {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
			}
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.runHelmCommand(testCase.helmParams)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedExecCalls, utils.Calls)
		})
	}
}
