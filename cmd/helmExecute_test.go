package cmd

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/SAP/jenkins-library/pkg/kubernetes/mocks"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperenv"
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

	cpe := helmExecuteCommonPipelineEnvironment{}
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

			err := runHelmExecute(testCase.config, helmExecute, &fileHandlerMock{}, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmLint(t *testing.T) {
	t.Parallel()

	cpe := helmExecuteCommonPipelineEnvironment{}
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

			err := runHelmExecute(testCase.config, helmExecute, &fileHandlerMock{}, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmInstall(t *testing.T) {
	t.Parallel()

	cpe := helmExecuteCommonPipelineEnvironment{}
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

			err := runHelmExecute(testCase.config, helmExecute, &fileHandlerMock{}, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmTest(t *testing.T) {
	t.Parallel()

	cpe := helmExecuteCommonPipelineEnvironment{}
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

			err := runHelmExecute(testCase.config, helmExecute, &fileHandlerMock{}, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmUninstall(t *testing.T) {
	t.Parallel()

	cpe := helmExecuteCommonPipelineEnvironment{}
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

			err := runHelmExecute(testCase.config, helmExecute, &fileHandlerMock{}, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmDependency(t *testing.T) {
	t.Parallel()

	cpe := helmExecuteCommonPipelineEnvironment{}
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

			err := runHelmExecute(testCase.config, helmExecute, &fileHandlerMock{}, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmPush(t *testing.T) {
	t.Parallel()

	cpe := helmExecuteCommonPipelineEnvironment{}
	testTable := []struct {
		config         helmExecuteOptions
		methodString   string
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "publish",
			},
			methodString: "https://my.target.repository",
			methodError:  nil,
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
			helmExecute.On("RunHelmPublish").Return(testCase.methodString, testCase.methodError)

			err := runHelmExecute(testCase.config, helmExecute, &fileHandlerMock{}, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmDefaultCommand(t *testing.T) {
	t.Parallel()

	cpe := helmExecuteCommonPipelineEnvironment{}
	testTable := []struct {
		config             helmExecuteOptions
		methodLintError    error
		methodPackageError error
		methodPublishError error
		expectedErrStr     string
		fileUtils          fileHandlerMock
		assertFunc         func(fileHandlerMock) error
	}{
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodLintError:    nil,
			methodPackageError: nil,
			methodPublishError: nil,
			fileUtils:          fileHandlerMock{},
		},
		{
			// this test checks if parseAndRenderCPETemplate is called properly
			// when config.RenderValuesTemplate is true
			config: helmExecuteOptions{
				HelmCommand:          "",
				RenderValuesTemplate: true,
			},
			methodLintError:    nil,
			methodPackageError: nil,
			methodPublishError: nil,
			fileUtils:          fileHandlerMock{},
			// we expect the values file is traversed since parsing and rendering according to cpe template is active
			assertFunc: func(f fileHandlerMock) error {
				if len(f.fileExistsCalled) == 1 && f.fileExistsCalled[0] == "/values.yaml" {
					return nil
				}
				return fmt.Errorf("expected FileExists called for ['/values.yaml'] but was: %+v", f.fileExistsCalled)
			},
		},
		{
			// this test checks if parseAndRenderCPETemplate is NOT called
			// when config.RenderValuesTemplate is false
			config: helmExecuteOptions{
				HelmCommand:          "",
				RenderValuesTemplate: false,
			},
			methodLintError:    nil,
			methodPackageError: nil,
			methodPublishError: nil,
			fileUtils:          fileHandlerMock{},
			// we expect the values file is not traversed since parsing and rendering according to cpe template is not active
			assertFunc: func(f fileHandlerMock) error {
				if len(f.fileExistsCalled) > 0 {
					return fmt.Errorf("expected FileExists not called, but was for: %+v", f.fileExistsCalled)
				}
				return nil
			},
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodLintError: errors.New("some error"),
			expectedErrStr:  "failed to execute helm lint: some error",
			fileUtils:       fileHandlerMock{},
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodPackageError: errors.New("some error"),
			expectedErrStr:     "failed to execute helm dependency: some error",
			fileUtils:          fileHandlerMock{},
		},
		{
			config: helmExecuteOptions{
				HelmCommand: "",
			},
			methodPublishError: errors.New("some error"),
			expectedErrStr:     "failed to execute helm publish: some error",
			fileUtils:          fileHandlerMock{},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecute := &mocks.HelmExecutor{}
			helmExecute.On("RunHelmDependency").Return(testCase.methodPackageError)
			helmExecute.On("RunHelmLint").Return(testCase.methodLintError)
			helmExecute.On("RunHelmPublish").Return(testCase.methodPublishError)

			err := runHelmExecute(testCase.config, helmExecute, &testCase.fileUtils, &cpe)
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
			if testCase.assertFunc != nil {
				assert.NoError(t, testCase.assertFunc(testCase.fileUtils))
			}

		})
	}

}

func TestParseAndRenderCPETemplate(t *testing.T) {
	commonPipelineEnvironment := "commonPipelineEnvironment"
	valuesYaml := []byte(`
image: "image_1"
tag: {{ cpe "artifactVersion" }}
`)
	values1Yaml := []byte(`
image: "image_2"
tag: {{ cpe "artVersion" }}
`)
	values3Yaml := []byte(`
image: "image_3"
tag: {{ .CPE.artVersion
`)
	values4Yaml := []byte(`
image: "test-image"
tag: {{ imageTag "test-image" }}
`)

	tmpDir := t.TempDir()
	require.DirExists(t, tmpDir)
	err := os.Mkdir(path.Join(tmpDir, commonPipelineEnvironment), 0700)
	require.NoError(t, err)
	cpe := piperenv.CPEMap{
		"artifactVersion":         "1.0.0-123456789",
		"container/imageNameTags": []string{"test-image:1.0.0-123456789"},
	}
	err = cpe.WriteToDisk(tmpDir)
	require.NoError(t, err)

	defaultValueFile := "values.yaml"
	config := helmExecuteOptions{
		ChartPath: ".",
	}

	tt := []struct {
		name             string
		defaultValueFile string
		config           helmExecuteOptions
		expectedErr      error
		valueFile        []byte
	}{
		{
			name:             "'artifactVersion' file exists in CPE",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      nil,
			valueFile:        valuesYaml,
		},
		{
			name:             "'artVersion' file does not exist in CPE",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      nil,
			valueFile:        values1Yaml,
		},
		{
			name:             "Good template ({{ imageTag 'test-image' }})",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      nil,
			valueFile:        values4Yaml,
		},
		{
			name:             "Wrong template ({{ .CPE.artVersion)",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      fmt.Errorf("failed to parse template: failed to parse cpe template '\nimage: \"image_3\"\ntag: {{ .CPE.artVersion\n': template: cpetemplate:4: unclosed action started at cpetemplate:3"),
			valueFile:        values3Yaml,
		},
		{
			name:             "Multiple value files",
			defaultValueFile: defaultValueFile,
			config: helmExecuteOptions{
				ChartPath:  ".",
				HelmValues: []string{"./values_1.yaml", "./values_2.yaml"},
			},
			expectedErr: nil,
			valueFile:   valuesYaml,
		},
		{
			name:             "No value file is provided",
			defaultValueFile: "",
			config: helmExecuteOptions{
				ChartPath:  ".",
				HelmValues: []string{},
			},
			expectedErr: fmt.Errorf("no value file to proccess, please provide value file(s)"),
			valueFile:   valuesYaml,
		},
		{
			name:             "Wrong path to value file",
			defaultValueFile: defaultValueFile,
			config: helmExecuteOptions{
				ChartPath:  ".",
				HelmValues: []string{"wrong/path/to/values_1.yaml"},
			},
			expectedErr: fmt.Errorf("failed to read file: could not read 'wrong/path/to/values_1.yaml'"),
			valueFile:   valuesYaml,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			utils := newHelmMockUtilsBundle()
			utils.AddFile(fmt.Sprintf("%s/%s", config.ChartPath, test.defaultValueFile), test.valueFile)

			if len(test.config.HelmValues) == 2 {
				for _, value := range test.config.HelmValues {
					utils.AddFile(value, test.valueFile)
				}
			}

			err := parseAndRenderCPETemplate(test.config, tmpDir, utils)
			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type fileHandlerMock struct {
	fileExistsCalled []string
	fileReadCalled   []string
	fileWriteCalled  []string
}

func (f *fileHandlerMock) FileWrite(name string, content []byte, mode os.FileMode) error {
	f.fileWriteCalled = append(f.fileWriteCalled, name)
	return nil
}

func (f *fileHandlerMock) FileRead(name string) ([]byte, error) {
	f.fileReadCalled = append(f.fileReadCalled, name)
	return []byte{}, nil
}

func (f *fileHandlerMock) FileExists(name string) (bool, error) {
	f.fileExistsCalled = append(f.fileExistsCalled, name)
	return true, nil
}
