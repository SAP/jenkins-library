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

	t.Run("Helm add command", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					ChartRepo: "https://charts.helm.sh/stable",
				},
				expectedConfig: []string{"repo", "add", "stable", "https://charts.helm.sh/stable"},
			},
		}

		for i, testCase := range testTable {
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmAdd()
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm upgrade command", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					DeploymentName:        "test_deployment",
					ChartPath:             ".",
					Namespace:             "test_namespace",
					ForceUpdates:          true,
					HelmDeployWaitSeconds: 3456,
					AdditionalParameters:  []string{"additional parameter"},
					ContainerRegistryURL:  "https://hub.docker.com/",
					Image:                 "dtzar/helm-kubectl:3.4.1",
				},
				expectedConfig: []string{"upgrade", "test_deployment", ".", "--install", "--namespace", "test_namespace", "--set", "image.repository=hub.docker.com/dtzar/helm-kubectl,image.tag=3.4.1", "--force", "--wait", "--timeout", "3456s", "--atomic", "additional parameter"},
			},
		}

		for i, testCase := range testTable {
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmUpgrade()
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm lint command", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					ChartPath: ".",
				},
				expectedConfig: []string{"lint", "."},
			},
		}

		for i, testCase := range testTable {
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmLint()
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm install command", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			config                HelmExecuteOptions
			expectedConfigInstall []string
			expectedConfigAdd     []string
		}{
			{
				config: HelmExecuteOptions{
					ChartPath:             ".",
					DeploymentName:        "testPackage",
					Namespace:             "test-namespace",
					HelmDeployWaitSeconds: 525,
					ChartRepo:             "https://charts.helm.sh/stable",
				},
				expectedConfigAdd:     []string{"repo", "add", "stable", "https://charts.helm.sh/stable"},
				expectedConfigInstall: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--wait", "--timeout", "525s"},
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
					ChartRepo:             "https://charts.helm.sh/stable",
				},
				expectedConfigAdd:     []string{"repo", "add", "stable", "https://charts.helm.sh/stable"},
				expectedConfigInstall: []string{"install", "testPackage", ".", "--namespace", "test-namespace", "--create-namespace", "--atomic", "--dry-run", "--wait", "--timeout", "525s", "--set-file my_script=dothings.sh"},
			},
		}

		for _, testCase := range testTable {
			utils := newHelmMockUtilsBundle()
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmInstall()
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfigAdd}, utils.Calls[0])
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfigInstall}, utils.Calls[1])
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
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmUninstall()
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

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
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmPackage()
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
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmTest()
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
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmUninstall()
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

		}
	})

	t.Run("Helm registry login command", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					HelmRegistryUser: "helmRegistryUser",
				},
				expectedConfig: []string{"registry login", "-u", "helmRegistryUser", "localhost:5000"},
			},
		}

		for i, testCase := range testTable {
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmRegistryLogin()
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[i])
		}
	})

	t.Run("Helm registry login command", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()

		testTable := []struct {
			config         HelmExecuteOptions
			expectedConfig []string
		}{
			{
				config: HelmExecuteOptions{
					DeploymentName: "nginx",
					PackageVersion: "2.3.4",
				},
				expectedConfig: []string{"push", "nginx2.3.4.tgz", "oci://localhost:5000/helm-charts"},
			},
		}

		for _, testCase := range testTable {
			helmExecute := HelmExecute{
				utils:   utils,
				config:  testCase.config,
				verbose: false,
				stdout:  log.Writer(),
			}
			err := helmExecute.RunHelmPush()
			assert.NoError(t, err)
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: testCase.expectedConfig}, utils.Calls[1])
		}
	})

}
