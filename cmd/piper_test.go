package cmd

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

type stepOptions struct {
	TestParam string `json:"testParam,omitempty"`
}

func openFileMock(name string) (io.ReadCloser, error) {
	var r string
	switch name {
	case "testDefaults.yml":
		r = "general:\n  testParam: testValue"
	case "testDefaultsInvalid.yml":
		r = "invalid yaml"
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
}

func TestPrepareConfig(t *testing.T) {
	defaultsBak := generalConfig.defaultConfig
	generalConfig.defaultConfig = []string{"testDefaults.yml"}
	defer func() { generalConfig.defaultConfig = defaultsBak }()

	t.Run("using stepConfigJSON", func(t *testing.T) {
		stepConfigJSONBak := generalConfig.stepConfigJSON
		generalConfig.stepConfigJSON = `{"testParam": "testValueJSON"}`
		defer func() { generalConfig.stepConfigJSON = stepConfigJSONBak }()
		testOptions := stepOptions{}
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

		PrepareConfig(testCmd, &metadata, "testStep", &testOptions, openFileMock)
		assert.Equal(t, "testValueJSON", testOptions.TestParam, "wrong value retrieved from config")
	})

	t.Run("using config files", func(t *testing.T) {
		t.Run("success case", func(t *testing.T) {
			testOptions := stepOptions{}
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

			err := PrepareConfig(testCmd, &metadata, "testStep", &testOptions, openFileMock)
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
			generalConfig.defaultConfig = []string{"testDefaultsInvalid.yml"}
			testOptions := stepOptions{}
			var testCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
			metadata := config.StepData{}

			err := PrepareConfig(testCmd, &metadata, "testStep", &testOptions, openFileMock)
			assert.Error(t, err, "error expected but none occured")
		})
	})
}
