package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/mock"
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
			assert.NoError(t, err, "no error expected but error occured")

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
			assert.Error(t, err, "error expected but none occured")
		})
	})
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
	})
	t.Run("Exits on unsupported type mismatch", func(t *testing.T) {
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
}
