package cmd

import (
	"fmt"
	"os"

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

	rootCmd.AddCommand(PiperGetConfig())
	rootCmd.PersistentFlags().StringVar(&generalConfig.customConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&generalConfig.defaultConfig, "defaultConfig", nil, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().StringVar(&generalConfig.parametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&generalConfig.stageName, "stageName", os.Getenv("STAGE_NAME"), "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&generalConfig.stepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&generalConfig.verbose, "verbose", "v", false, "verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
