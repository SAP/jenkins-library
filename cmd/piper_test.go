package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestAddRootFlags(t *testing.T) {
	var testRootCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
	addRootFlags(testRootCmd)

	assert.NotNil(t, testRootCmd.Flag("customConfig"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("defaultConfig"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("parametersJSON"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("stageName"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("stepConfigJSON"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("verbose"), "expected flag not available")

}

func TestAdoptStageNameFromParametersJSON(t *testing.T) {
	tt := []struct {
		name          string
		stageNameArg  string
		stageNameEnv  string
		stageNameJSON string
	}{
		{name: "no stage name", stageNameArg: "", stageNameEnv: "", stageNameJSON: ""},
		{name: "stage name arg+env", stageNameArg: "arg", stageNameEnv: "env", stageNameJSON: "json"},
		{name: "stage name env", stageNameArg: "", stageNameEnv: "env", stageNameJSON: "json"},
		{name: "stage name json", stageNameArg: "", stageNameEnv: "", stageNameJSON: "json"},
		{name: "stage name arg", stageNameArg: "arg", stageNameEnv: "", stageNameJSON: "json"},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			// init
			GeneralConfig.StageName = test.stageNameArg

			resetValue := os.Getenv(stageNameEnvKey)
			defer func() { _ = os.Setenv(stageNameEnvKey, resetValue) }()

			err := os.Setenv(stageNameEnvKey, test.stageNameEnv)
			if err != nil {
				t.Fatalf("could not set env var %s", stageNameEnvKey)
			}

			if test.stageNameJSON != "" {
				GeneralConfig.ParametersJSON = fmt.Sprintf("{\"stageName\":\"%s\"}", test.stageNameJSON)
			} else {
				GeneralConfig.ParametersJSON = "{}"
			}
			// test
			initStageName(false)

			// assert
			// Order of if-clauses reflects wanted precedence.
			if test.stageNameArg != "" {
				assert.Equal(t, test.stageNameArg, GeneralConfig.StageName)
			} else if test.stageNameJSON != "" {
				assert.Equal(t, test.stageNameJSON, GeneralConfig.StageName)
			} else if test.stageNameEnv != "" {
				assert.Equal(t, test.stageNameEnv, GeneralConfig.StageName)
			} else {
				assert.Equal(t, "", GeneralConfig.StageName)
			}
		})
	}
}

func TestPrepareConfig(t *testing.T) {
	defaultsBak := GeneralConfig.DefaultConfig
	GeneralConfig.DefaultConfig = []string{"testDefaults.yml"}
	defer func() { GeneralConfig.DefaultConfig = defaultsBak }()

	t.Run("using stepConfigJSON", func(t *testing.T) {
		stepConfigJSONBak := GeneralConfig.StepConfigJSON
		GeneralConfig.StepConfigJSON = `{"testParam": "testValueJSON"}`
		defer func() { GeneralConfig.StepConfigJSON = stepConfigJSONBak }()
		testOptions := mock.StepOptions{}
		var testCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
		testCmd.Flags().StringVar(&testOptions.TestParam, "testParam", "", "test usage")
		metadata := config.StepData{
			Spec: config.StepSpec{
				Inputs: config.StepInputs{
					Parameters: []config.StepParameters{
						{Name: "testParam", Scope: []string{"GENERAL"}},
					},
				},
			},
		}

		PrepareConfig(testCmd, &metadata, "testStep", &testOptions, mock.OpenFileMock)
		assert.Equal(t, "testValueJSON", testOptions.TestParam, "wrong value retrieved from config")
	})

	t.Run("using config files", func(t *testing.T) {
		t.Run("success case", func(t *testing.T) {
			testOptions := mock.StepOptions{}
			var testCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
			testCmd.Flags().StringVar(&testOptions.TestParam, "testParam", "", "test usage")
			metadata := config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testParam", Scope: []string{"GENERAL"}},
						},
					},
				},
			}

			err := PrepareConfig(testCmd, &metadata, "testStep", &testOptions, mock.OpenFileMock)
			assert.NoError(t, err, "no error expected but error occurred")

			//assert config
			assert.Equal(t, "testValue", testOptions.TestParam, "wrong value retrieved from config")

			//assert that flag has been marked as changed
			testCmd.Flags().VisitAll(func(pflag *flag.Flag) {
				if pflag.Name == "testParam" {
					assert.True(t, pflag.Changed, "flag should be marked as changed")
				}
			})
		})

		t.Run("error case", func(t *testing.T) {
			GeneralConfig.DefaultConfig = []string{"testDefaultsInvalid.yml"}
			testOptions := mock.StepOptions{}
			var testCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
			metadata := config.StepData{}

			err := PrepareConfig(testCmd, &metadata, "testStep", &testOptions, mock.OpenFileMock)
			assert.Error(t, err, "error expected but none occurred")
		})
	})
}

func TestRetrieveHookConfig(t *testing.T) {
	tt := []struct {
		hookJSON           []byte
		expectedHookConfig HookConfiguration
	}{
		{hookJSON: []byte(""), expectedHookConfig: HookConfiguration{}},
		{hookJSON: []byte(`{"sentry":{"dsn":"https://my.sentry.dsn"}}`), expectedHookConfig: HookConfiguration{SentryConfig: SentryConfiguration{Dsn: "https://my.sentry.dsn"}}},
	}

	for _, test := range tt {
		var target HookConfiguration
		var hookJSONRaw json.RawMessage
		if len(test.hookJSON) > 0 {
			err := json.Unmarshal(test.hookJSON, &hookJSONRaw)
			assert.NoError(t, err)
		}
		retrieveHookConfig(&hookJSONRaw, &target)
		assert.Equal(t, test.expectedHookConfig, target)
	}
}

func TestGetProjectConfigFile(t *testing.T) {

	tt := []struct {
		filename       string
		filesAvailable []string
		expected       string
	}{
		{filename: ".pipeline/config.yml", filesAvailable: []string{}, expected: ".pipeline/config.yml"},
		{filename: ".pipeline/config.yml", filesAvailable: []string{".pipeline/config.yml"}, expected: ".pipeline/config.yml"},
		{filename: ".pipeline/config.yml", filesAvailable: []string{".pipeline/config.yaml"}, expected: ".pipeline/config.yaml"},
		{filename: ".pipeline/config.yaml", filesAvailable: []string{".pipeline/config.yml", ".pipeline/config.yaml"}, expected: ".pipeline/config.yaml"},
		{filename: ".pipeline/config.yml", filesAvailable: []string{".pipeline/config.yml", ".pipeline/config.yaml"}, expected: ".pipeline/config.yml"},
	}

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			dir, err := ioutil.TempDir("", "")
			defer os.RemoveAll(dir) // clean up
			assert.NoError(t, err)

			if len(test.filesAvailable) > 0 {
				configFolder := filepath.Join(dir, filepath.Dir(test.filesAvailable[0]))
				err = os.MkdirAll(configFolder, 0700)
				assert.NoError(t, err)
			}

			for _, file := range test.filesAvailable {
				ioutil.WriteFile(filepath.Join(dir, file), []byte("general:"), 0700)
			}

			assert.Equal(t, filepath.Join(dir, test.expected), getProjectConfigFile(filepath.Join(dir, test.filename)))
		})
	}
}

func TestConvertTypes(t *testing.T) {
	t.Run("Converts strings to booleans", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		options := struct {
			Foo bool `json:"foo,omitempty"`
			Bar bool `json:"bar,omitempty"`
		}{}
		options.Foo = true
		options.Bar = false

		stepConfig := map[string]interface{}{}
		stepConfig["foo"] = "False"
		stepConfig["bar"] = "True"

		// Test
		stepConfig = checkTypes(stepConfig, options)

		confJSON, _ := json.Marshal(stepConfig)
		_ = json.Unmarshal(confJSON, &options)

		// Assert
		assert.Equal(t, false, stepConfig["foo"])
		assert.Equal(t, true, stepConfig["bar"])
		assert.Equal(t, false, options.Foo)
		assert.Equal(t, true, options.Bar)
		assert.False(t, hasFailed, "Expected checkTypes() NOT to exit via logging framework")
	})
	t.Run("Converts numbers to strings", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		options := struct {
			Foo string `json:"foo,omitempty"`
			Bar string `json:"bar,omitempty"`
		}{}

		stepConfig := map[string]interface{}{}
		stepConfig["foo"] = 1.5
		stepConfig["bar"] = 42

		// Test
		stepConfig = checkTypes(stepConfig, options)

		confJSON, _ := json.Marshal(stepConfig)
		_ = json.Unmarshal(confJSON, &options)

		// Assert
		assert.Equal(t, "1.5", stepConfig["foo"])
		assert.Equal(t, "42", stepConfig["bar"])
		assert.Equal(t, "1.5", options.Foo)
		assert.Equal(t, "42", options.Bar)
		assert.False(t, hasFailed, "Expected checkTypes() NOT to exit via logging framework")
	})
	t.Run("Keeps numbers", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		options := struct {
			Foo int     `json:"foo,omitempty"`
			Bar float32 `json:"bar,omitempty"`
		}{}

		stepConfig := map[string]interface{}{}

		content := []byte(`
foo: 1
bar: 42
`)
		err := yaml.Unmarshal(content, &stepConfig)
		assert.NoError(t, err)

		// Test
		stepConfig = checkTypes(stepConfig, options)

		confJSON, _ := json.Marshal(stepConfig)
		_ = json.Unmarshal(confJSON, &options)

		// Assert
		assert.Equal(t, 1, stepConfig["foo"])
		assert.Equal(t, float32(42.0), stepConfig["bar"])
		assert.Equal(t, 1, options.Foo)
		assert.Equal(t, float32(42.0), options.Bar)
		assert.False(t, hasFailed, "Expected checkTypes() NOT to exit via logging framework")
	})
	t.Run("Exits because string found, slice expected", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		options := struct {
			Foo []string `json:"foo,omitempty"`
		}{}

		stepConfig := map[string]interface{}{}
		stepConfig["foo"] = "entry"

		// Test
		stepConfig = checkTypes(stepConfig, options)

		// Assert
		assert.True(t, hasFailed, "Expected checkTypes() to exit via logging framework")
	})
	t.Run("Exits because float found, int expected", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		options := struct {
			Foo int `json:"foo,omitempty"`
		}{}

		stepConfig := map[string]interface{}{}

		content := []byte("foo: 1.1")
		err := yaml.Unmarshal(content, &stepConfig)
		assert.NoError(t, err)

		// Test
		stepConfig = checkTypes(stepConfig, options)

		// Assert
		assert.Equal(t, 1.1, stepConfig["foo"])
		assert.True(t, hasFailed, "Expected checkTypes() to exit via logging framework")
	})
	t.Run("Exits in case number beyond length", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		options := struct {
			Foo string `json:"foo,omitempty"`
		}{}

		stepConfig := map[string]interface{}{}

		content := []byte("foo: 73554900100200011600")
		err := yaml.Unmarshal(content, &stepConfig)
		assert.NoError(t, err)

		// Test
		stepConfig = checkTypes(stepConfig, options)

		// Assert
		assert.True(t, hasFailed, "Expected checkTypes() to exit via logging framework")
	})
	t.Run("Ignores nil values", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		options := struct {
			Foo []string `json:"foo,omitempty"`
			Bar string   `json:"bar,omitempty"`
		}{}

		stepConfig := map[string]interface{}{}
		stepConfig["foo"] = nil
		stepConfig["bar"] = nil

		// Test
		stepConfig = checkTypes(stepConfig, options)
		confJSON, _ := json.Marshal(stepConfig)
		_ = json.Unmarshal(confJSON, &options)

		// Assert
		assert.Nil(t, stepConfig["foo"])
		assert.Nil(t, stepConfig["bar"])
		assert.Equal(t, []string(nil), options.Foo)
		assert.Equal(t, "", options.Bar)
		assert.False(t, hasFailed, "Expected checkTypes() NOT to exit via logging framework")
	})
	t.Run("Logs warning for unknown type-mismatches", func(t *testing.T) {
		// Init
		hasFailed := false

		exitFunc := log.Entry().Logger.ExitFunc
		log.Entry().Logger.ExitFunc = func(int) {
			hasFailed = true
		}
		defer func() { log.Entry().Logger.ExitFunc = exitFunc }()

		logBuffer := new(bytes.Buffer)

		logOutput := log.Entry().Logger.Out
		log.Entry().Logger.Out = logBuffer
		defer func() { log.Entry().Logger.Out = logOutput }()

		options := struct {
			Foo string `json:"foo,omitempty"`
		}{}

		stepConfig := map[string]interface{}{}
		stepConfig["foo"] = true

		// Test
		stepConfig = checkTypes(stepConfig, options)
		confJSON, _ := json.Marshal(stepConfig)
		_ = json.Unmarshal(confJSON, &options)

		// Assert
		assert.Equal(t, true, stepConfig["foo"])
		assert.Equal(t, "", options.Foo)
		assert.Contains(t, logBuffer.String(), "The value may be ignored as a result")
		assert.False(t, hasFailed, "Expected checkTypes() NOT to exit via logging framework")
	})
}
