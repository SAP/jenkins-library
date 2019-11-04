package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type generalConfigOptions struct {
	customConfig   string
	defaultConfig  []string //ordered list of Piper default configurations. Can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	parametersJSON string
	stageName      string
	stepConfigJSON string
	stepMetadata   string //metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	stepName       string
	verbose        bool
}

var rootCmd = &cobra.Command{
	Use:   "piper",
	Short: "Executes CI/CD steps from project 'Piper' ",
	Long: `
This project 'Piper' binary provides a CI/CD step libary.
It contains many steps which can be used within CI/CD systems as well as directly on e.g. a developer's machine.
`,
	//ToDo: respect stageName to also come from parametersJSON -> first env.STAGE_NAME, second: parametersJSON, third: flag
}

var generalConfig generalConfigOptions

// Execute is the starting point of the piper command line tool
func Execute() {

	rootCmd.AddCommand(ConfigCommand())
	rootCmd.AddCommand(KarmaExecuteTestsCommand())
	rootCmd.AddCommand(GithubPublishReleaseCommand())

	addRootFlags(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addRootFlags(rootCmd *cobra.Command) {

	rootCmd.PersistentFlags().StringVar(&generalConfig.customConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&generalConfig.defaultConfig, "defaultConfig", nil, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().StringVar(&generalConfig.parametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&generalConfig.stageName, "stageName", os.Getenv("STAGE_NAME"), "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&generalConfig.stepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&generalConfig.verbose, "verbose", "v", false, "verbose output")

}

// PrepareConfig reads step configuration from various sources and merges it (defaults, config file, flags, ...)
func PrepareConfig(cmd *cobra.Command, metadata *config.StepData, stepName string, options interface{}, openFile func(s string) (io.ReadCloser, error)) error {

	filters := metadata.GetParameterFilters()

	flagValues := config.AvailableFlagValues(cmd, &filters)

	var myConfig config.Config
	var stepConfig config.StepConfig

	if len(generalConfig.stepConfigJSON) != 0 {
		// ignore config & defaults in favor of passed stepConfigJSON
		stepConfig = config.GetStepConfigWithJSON(flagValues, generalConfig.stepConfigJSON, filters)
	} else {
		// use config & defaults

		//accept that config file and defaults cannot be loaded since both are not mandatory here
		customConfig, _ := openFile(generalConfig.customConfig)
		var defaultConfig []io.ReadCloser
		for _, f := range generalConfig.defaultConfig {
			//ToDo: support also https as source
			fc, _ := openFile(f)
			defaultConfig = append(defaultConfig, fc)
		}

		var err error
		stepConfig, err = myConfig.GetStepConfig(flagValues, generalConfig.parametersJSON, customConfig, defaultConfig, filters, metadata.Spec.Inputs.Parameters, generalConfig.stageName, stepName)
		if err != nil {
			return errors.Wrap(err, "retrieving step configuration failed")
		}
	}

	confJSON, _ := json.Marshal(stepConfig.Config)
	json.Unmarshal(confJSON, &options)

	config.MarkFlagsWithValue(cmd, stepConfig)

	return nil
}

func openPiperFile(name string) (io.ReadCloser, error) {
	//ToDo: support also https as source
	if !strings.HasPrefix(name, "http") {
		return os.Open(name)
	}
	return nil, fmt.Errorf("file location not yet supported for '%v'", name)
}
