package cmd

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/SAP/jenkins-library/pkg/kubernetes/mocks"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type helmMockUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
}

func newHelmMockUtilsBundle() helmMockUtilsBundle {
	utils := helmMockUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		HttpClientMock: &mock.HttpClientMock{
			FileUploads: map[string]string{},
		},
	}
	return utils
}

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

func TestRunHelmDependency(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		config         helmExecuteOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "dependency",
			},
			methodError: nil,
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "dependency",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm dependency: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmDependency").Return(testCase.methodError)

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
			expectedErrStr:     "failed to execute helm dependency: some error",
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
			helmExecute.On("RunHelmDependency").Return(testCase.methodPackageError)
			helmExecute.On("RunHelmPublish").Return(testCase.methodPublishError)

			err := runHelmExecute(testCase.config, helmExecute)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}

		})
	}

}

//TODO: implement Table-Driven Testing
func TestGetAndRenderImageInfo(t *testing.T) {
	commonPipelineEnvironment := "commonPipelineEnvironment"
	valuesYaml := []byte(`
image: "image_1"
tag: {{ .CPE.artifactVersion }}
`)
	values1Yaml := []byte(`
image: "image_2"
tag: {{ .CPE.artVersion }}
`)
	values3Yaml := []byte(`
image: "image_3"
tag: {{ .CPE.artVersion
`)

	tmpDir, err := os.MkdirTemp(os.TempDir(), "test-data-*")
	require.NoError(t, err)
	err = os.Mkdir(path.Join(tmpDir, commonPipelineEnvironment), 0700)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(tmpDir, commonPipelineEnvironment, "artifactVersion"), []byte("1.0.0-123456789"), 0700)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	config := helmExecuteOptions{
		ChartPath: ".",
	}

	t.Run("'artifactVersion' file exists in CPE", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()
		utils.AddFile(fmt.Sprintf("%s/%s", config.ChartPath, "values.yaml"), valuesYaml)

		err = getAndRenderImageInfo(config, tmpDir, utils)
		assert.NoError(t, err)
	})

	t.Run("'artVersion' file does not exist in CPE", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()
		utils.AddFile(fmt.Sprintf("%s/%s", config.ChartPath, "values.yaml"), values1Yaml)

		err = getAndRenderImageInfo(config, tmpDir, utils)
		assert.NoError(t, err)
	})

	t.Run("Wrong template {{ .CPE.artVersion", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()
		utils.AddFile(fmt.Sprintf("%s/%s", config.ChartPath, "values.yaml"), values3Yaml)

		err = getAndRenderImageInfo(config, tmpDir, utils)
		assert.EqualError(t, err, "failed to parse template: template: new:4: unclosed action started at new:3")
	})

	t.Run("Multiple values files", func(t *testing.T) {
		config.HelmValues = []string{"./values_1.yaml", "./values_2.yaml"}

		utils := newHelmMockUtilsBundle()
		utils.AddFile(fmt.Sprintf("%s/%s", config.ChartPath, "values.yaml"), valuesYaml)
		utils.AddFile(config.HelmValues[0], valuesYaml)
		utils.AddFile(config.HelmValues[1], valuesYaml)

		err = getAndRenderImageInfo(config, tmpDir, utils)
		assert.NoError(t, err)
	})

	t.Run("Wrong path to values file", func(t *testing.T) {
		config.HelmValues = []string{"wrong/path/to/values_1.yaml"}

		utils := newHelmMockUtilsBundle()
		utils.AddFile(fmt.Sprintf("%s/%s", config.ChartPath, "values.yaml"), valuesYaml)

		err = getAndRenderImageInfo(config, tmpDir, utils)
		assert.EqualError(t, err, "failed to read file: could not read 'wrong/path/to/values_1.yaml'")
	})
}
