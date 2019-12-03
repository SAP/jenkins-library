package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// GeneralConfigOptions contains all global configuration options for piper binary
type GeneralConfigOptions struct {
	CustomConfig   string
	DefaultConfig  []string //ordered list of Piper default configurations. Can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	ParametersJSON string
	StageName      string
	StepConfigJSON string
	StepMetadata   string //metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	StepName       string
	Verbose        bool
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

// GeneralConfig contains global configuration flags for piper binary
var GeneralConfig GeneralConfigOptions

// Execute is the starting point of the piper command line tool
func Execute() {

	rootCmd.AddCommand(ConfigCommand())
	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(KarmaExecuteTestsCommand())
	rootCmd.AddCommand(GithubPublishReleaseCommand())

	addRootFlags(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addRootFlags(rootCmd *cobra.Command) {

	rootCmd.PersistentFlags().StringVar(&GeneralConfig.CustomConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&GeneralConfig.DefaultConfig, "defaultConfig", []string{".pipeline/defaults.yaml"}, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.ParametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StageName, "stageName", os.Getenv("STAGE_NAME"), "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&GeneralConfig.Verbose, "verbose", "v", false, "verbose output")

}

// PrepareConfig reads step configuration from various sources and merges it (defaults, config file, flags, ...)
func PrepareConfig(cmd *cobra.Command, metadata *config.StepData, stepName string, options interface{}, openFile func(s string) (io.ReadCloser, error)) error {

	filters := metadata.GetParameterFilters()

	flagValues := config.AvailableFlagValues(cmd, &filters)

	var myConfig config.Config
	var stepConfig config.StepConfig

	if len(GeneralConfig.StepConfigJSON) != 0 {
		// ignore config & defaults in favor of passed stepConfigJSON
		stepConfig = config.GetStepConfigWithJSON(flagValues, GeneralConfig.StepConfigJSON, filters)
	} else {
		// use config & defaults
		var customConfig io.ReadCloser
		var err error
		//accept that config file and defaults cannot be loaded since both are not mandatory here
		if projectConfigFile, err := getProjectConfigFile(GeneralConfig.CustomConfig); err != nil {
			return err
		} else {
			if piperutils.FileExists(projectConfigFile) {
				if customConfig, err = openFile(projectConfigFile); err != nil {
					errors.Wrapf(err, "Cannot read '%s'", projectConfigFile)
				}
			} else {
				log.Entry().Infof("Project config file '%s' does not exist. No project configuration available.", projectConfigFile)
				customConfig = nil
			}
		}

		var defaultConfig []io.ReadCloser
		for _, f := range GeneralConfig.DefaultConfig {
			//ToDo: support also https as source
			fc, _ := openFile(f)
			defaultConfig = append(defaultConfig, fc)
		}

		stepConfig, err = myConfig.GetStepConfig(flagValues, GeneralConfig.ParametersJSON, customConfig, defaultConfig, filters, metadata.Spec.Inputs.Parameters, GeneralConfig.StageName, stepName)
		if err != nil {
			return errors.Wrap(err, "retrieving step configuration failed")
		}
	}

	confJSON, _ := json.Marshal(stepConfig.Config)
	json.Unmarshal(confJSON, &options)

	config.MarkFlagsWithValue(cmd, stepConfig)

	return nil
}

func getProjectConfigFile(configured string) (string, error) {

	configFolder := ".pipeline"
	defaultConfigFiles := []string{fmt.Sprintf("%s/%s", configFolder, "config.yml"), fmt.Sprintf("%s/%s", configFolder, "config.yaml")}

	explicitlyConfigured := ! contains(defaultConfigFiles, configured)

        if( explicitlyConfigured) {
		return configured, nil
	}

	ymlExists := piperutils.FileExists(defaultConfigFiles[0])
	yamlExists := piperutils.FileExists(defaultConfigFiles[1])

	if ymlExists && yamlExists {
		return "", errors.New(fmt.Sprintf("'%s' and '%s' exists at the same time, can't judge which to use",
		defaultConfigFiles[0], defaultConfigFiles[1]))
	}

	if yamlExists {
		return defaultConfigFiles[1], nil
	}

	// we return this file also in case it does not exist. Later in the code flow it needs to be checked if
	// that file exists. Here we derive only which one should be used in general.
	return defaultConfigFiles[0], nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
