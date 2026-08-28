//go:build unit
// +build unit

package kubernetes

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				{Exec: "helm", Params: []string{"uninstall", "testPackage", "--namespace", "test-namespace", "--wait", "--timeout", "524s", "--debug", "--dry-run"}},
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

func TestRunHelmTemplate(t *testing.T) {
	renderedManifests := "apiVersion: apps/v1\nkind: Deployment\nspec:\n  template:\n    spec:\n      containers:\n        - image: registry.example.com/app:1.2.3\n"

	t.Run("argument construction and stdout capture", func(t *testing.T) {
		testTable := []struct {
			name           string
			config         HelmExecuteOptions
			verbose        bool
			expectedParams []string
		}{
			{
				name:           "minimal",
				config:         HelmExecuteOptions{ChartPath: ".", DeploymentName: "testChart"},
				expectedParams: []string{"template", "testChart", "."},
			},
			{
				name:           "with values files",
				config:         HelmExecuteOptions{ChartPath: ".", DeploymentName: "testChart", HelmValues: []string{"a.yaml", "b.yaml"}},
				expectedParams: []string{"template", "testChart", ".", "--values", "a.yaml", "--values", "b.yaml"},
			},
			{
				name:           "with namespace",
				config:         HelmExecuteOptions{ChartPath: ".", DeploymentName: "testChart", Namespace: "prod"},
				expectedParams: []string{"template", "testChart", ".", "--namespace", "prod"},
			},
			{
				name:           "verbose adds --debug",
				config:         HelmExecuteOptions{ChartPath: ".", DeploymentName: "testChart"},
				verbose:        true,
				expectedParams: []string{"template", "testChart", ".", "--debug"},
			},
			{
				name:           "all options combined",
				config:         HelmExecuteOptions{ChartPath: "charts/app", DeploymentName: "testChart", HelmValues: []string{"v.yaml"}, Namespace: "ns"},
				verbose:        true,
				expectedParams: []string{"template", "testChart", "charts/app", "--values", "v.yaml", "--namespace", "ns", "--debug"},
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				utils := helmMockUtilsBundle{
					ExecMockRunner: &mock.ExecMockRunner{
						StdoutReturn: map[string]string{"helm template.*": renderedManifests},
					},
				}
				helmExecute := HelmExecute{
					utils:   utils,
					config:  testCase.config,
					verbose: testCase.verbose,
					stdout:  log.Writer(),
				}

				out, err := helmExecute.RunHelmTemplate()

				require.NoError(t, err)
				require.Len(t, utils.Calls, 1)
				assert.Equal(t, "helm", utils.Calls[0].Exec)
				assert.Equal(t, testCase.expectedParams, utils.Calls[0].Params)
				assert.Equal(t, renderedManifests, string(out), "must return the manifests captured from stdout")
			})
		}
	})

	t.Run("missing ChartPath returns an error, no exec", func(t *testing.T) {
		utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
		helmExecute := HelmExecute{utils: utils, config: HelmExecuteOptions{DeploymentName: "testChart"}, stdout: log.Writer()}

		out, err := helmExecute.RunHelmTemplate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "chartPath value is mandatory")
		assert.Nil(t, out)
		assert.Empty(t, utils.Calls, "helm must not be invoked when ChartPath is missing")
	})

	t.Run("helm execution failure is surfaced", func(t *testing.T) {
		utils := helmMockUtilsBundle{
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"helm template.*": fmt.Errorf("template boom")},
			},
		}
		helmExecute := HelmExecute{utils: utils, config: HelmExecuteOptions{ChartPath: ".", DeploymentName: "testChart"}, stdout: log.Writer()}

		out, err := helmExecute.RunHelmTemplate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute helm template")
		assert.Contains(t, err.Error(), "template boom")
		assert.Nil(t, out)
	})

	t.Run("stdout is restored after the call", func(t *testing.T) {
		utils := helmMockUtilsBundle{
			ExecMockRunner: &mock.ExecMockRunner{
				StdoutReturn: map[string]string{"helm template.*": renderedManifests},
			},
		}
		originalStdout := log.Writer()
		helmExecute := HelmExecute{utils: utils, config: HelmExecuteOptions{ChartPath: ".", DeploymentName: "testChart"}, stdout: originalStdout}

		_, err := helmExecute.RunHelmTemplate()

		require.NoError(t, err)
		assert.Equal(t, originalStdout, utils.GetStdout(), "the capture buffer must be replaced by the original stdout after the call")
	})
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

func TestRunHelmPackageSigning(t *testing.T) {
	t.Run("signing flags are passed when both signingKey and signingKeyRing are set", func(t *testing.T) {
		utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
		helmExecute := HelmExecute{
			utils: utils,
			config: HelmExecuteOptions{
				ChartPath:      ".",
				SigningKey:     "My Signing Key <key@example.com>",
				SigningKeyRing: "/tmp/keyring.gpg",
			},
			stdout: log.Writer(),
		}

		err := helmExecute.runHelmPackage()

		require.NoError(t, err)
		require.Len(t, utils.Calls, 1)
		assert.Equal(t, "helm", utils.Calls[0].Exec)
		assert.Contains(t, utils.Calls[0].Params, "--sign")
		assert.Contains(t, utils.Calls[0].Params, "--key")
		assert.Contains(t, utils.Calls[0].Params, "My Signing Key <key@example.com>")
		assert.Contains(t, utils.Calls[0].Params, "--keyring")
		assert.Contains(t, utils.Calls[0].Params, "/tmp/keyring.gpg")
	})

	t.Run("no signing flags when both signingKey and signingKeyRing are empty", func(t *testing.T) {
		utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
		helmExecute := HelmExecute{
			utils:  utils,
			config: HelmExecuteOptions{ChartPath: "."},
			stdout: log.Writer(),
		}

		err := helmExecute.runHelmPackage()

		require.NoError(t, err)
		assert.NotContains(t, utils.Calls[0].Params, "--sign")
	})

	t.Run("error when signingKey is set but signingKeyRing is missing", func(t *testing.T) {
		utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
		helmExecute := HelmExecute{
			utils: utils,
			config: HelmExecuteOptions{
				ChartPath:  ".",
				SigningKey: "My Signing Key",
			},
			stdout: log.Writer(),
		}

		err := helmExecute.runHelmPackage()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "signingKey is set but signingKeyRing is missing")
		assert.Empty(t, utils.Calls, "helm must not be invoked when signing config is incomplete")
	})

	t.Run("error when signingKeyRing is set but signingKey is missing", func(t *testing.T) {
		utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
		helmExecute := HelmExecute{
			utils: utils,
			config: HelmExecuteOptions{
				ChartPath:      ".",
				SigningKeyRing: "/tmp/keyring.gpg",
			},
			stdout: log.Writer(),
		}

		err := helmExecute.runHelmPackage()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "signingKeyRing is set but signingKey is missing")
		assert.Empty(t, utils.Calls, "helm must not be invoked when signing config is incomplete")
	})
}

func TestRunHelmPublishSigning(t *testing.T) {
	t.Run("uploads both .tgz and .prov when signing is enabled", func(t *testing.T) {
		utils := helmMockUtilsBundle{
			ExecMockRunner: &mock.ExecMockRunner{},
			HttpClientMock: &mock.HttpClientMock{FileUploads: map[string]string{}},
		}
		utils.ReturnFileUploadStatus = 200

		helmExecute := HelmExecute{
			utils: utils,
			config: HelmExecuteOptions{
				TargetRepositoryURL:      "https://repo.example.com/",
				TargetRepositoryUser:     "user",
				TargetRepositoryPassword: "pass",
				PublishVersion:           "1.0.0",
				DeploymentName:           "mychart",
				ChartPath:                ".",
				SigningKey:               "My Key",
				SigningKeyRing:           "/tmp/keyring.gpg",
			},
			stdout: log.Writer(),
		}

		targetURL, err := helmExecute.RunHelmPublish()

		require.NoError(t, err)
		assert.Equal(t, "https://repo.example.com/mychart-1.0.0.tgz", targetURL)
		assert.Len(t, utils.FileUploads, 2)
		assert.Equal(t, "https://repo.example.com/mychart-1.0.0.tgz", utils.FileUploads["mychart-1.0.0.tgz"])
		assert.Equal(t, "https://repo.example.com/mychart-1.0.0.tgz.prov", utils.FileUploads["mychart-1.0.0.tgz.prov"])
	})

	t.Run("uploads only .tgz when signing is disabled", func(t *testing.T) {
		utils := helmMockUtilsBundle{
			ExecMockRunner: &mock.ExecMockRunner{},
			HttpClientMock: &mock.HttpClientMock{FileUploads: map[string]string{}},
		}
		utils.ReturnFileUploadStatus = 200

		helmExecute := HelmExecute{
			utils: utils,
			config: HelmExecuteOptions{
				TargetRepositoryURL:      "https://repo.example.com/",
				TargetRepositoryUser:     "user",
				TargetRepositoryPassword: "pass",
				PublishVersion:           "1.0.0",
				DeploymentName:           "mychart",
				ChartPath:                ".",
			},
			stdout: log.Writer(),
		}

		targetURL, err := helmExecute.RunHelmPublish()

		require.NoError(t, err)
		assert.Equal(t, "https://repo.example.com/mychart-1.0.0.tgz", targetURL)
		assert.Len(t, utils.FileUploads, 1)
		_, hasProv := utils.FileUploads["mychart-1.0.0.tgz.prov"]
		assert.False(t, hasProv, ".prov must not be uploaded when signing is disabled")
	})

	t.Run("error when .prov upload fails", func(t *testing.T) {
		utils := helmMockUtilsBundle{
			ExecMockRunner: &mock.ExecMockRunner{},
			HttpClientMock: &mock.HttpClientMock{
				FileUploads:           map[string]string{},
				ReturnFileUploadError: errors.New("connection reset"),
			},
		}

		helmExecute := HelmExecute{
			utils: utils,
			config: HelmExecuteOptions{
				TargetRepositoryURL: "https://repo.example.com/",
				PublishVersion:      "1.0.0",
				DeploymentName:      "mychart",
				ChartPath:           ".",
				SigningKey:          "My Key",
				SigningKeyRing:      "/tmp/keyring.gpg",
			},
			stdout: log.Writer(),
		}

		_, err := helmExecute.RunHelmPublish()

		require.Error(t, err)
	})
}
