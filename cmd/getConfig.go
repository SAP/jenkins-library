package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/config"
)

type PiperGetConfigOptions struct {
	output         string //output format, so far only JSON
	parametersJSON string //parameters to be considered in JSON format
	stepMetadata   string //metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	stepName       string
	contextConfig  bool
}

var configOptions PiperGetConfigOptions
var stepConfig config.StepConfig

// OpenFile defines the function to open files locally and remotely
var OpenFile = openPiperFile

// GetConfig is the entry command for loading the configuration of a pipeline step
func PiperGetConfig() *cobra.Command {
	var createConfigCmd = &cobra.Command{
		Use:   "getConfig",
		Short: "Loads the project 'Piper' configuration respecting defaults and parameters.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return generateConfig()
		},
	}

	addConfigFlags(createConfigCmd)
	return createConfigCmd
}

func generateConfig() error {

	var myConfig config.Config

	var metadata config.StepData
	metadataFile, err := OpenFile(configOptions.stepMetadata)
	if err != nil {
		return errors.Wrap(err, "metadata: open failed")
	}

	err = metadata.ReadPipelineStepData(metadataFile)
	if err != nil {
		return errors.Wrap(err, "metadata: read failed")
	}

	customConfig, err := OpenFile(generalConfig.customConfig)
	if err != nil {
		return errors.Wrap(err, "config: open failed")
	}

	defaultConfig, paramFilter, err := defaultsAndFilters(&metadata)
	if err != nil {
		return errors.Wrap(err, "defaults: retrieving step defaults failed")
	}

	for _, f := range generalConfig.defaultConfig {
		fc, err := OpenFile(f)
		if err != nil {
			return errors.Wrapf(err, "config: getting defaults failed: '%v'", f)
		}
		defaultConfig = append(defaultConfig, fc)
	}

	var flags map[string]interface{}

	stepConfig = myConfig.GetStepConfig(flags, generalConfig.parametersJSON, customConfig, defaultConfig, paramFilter, generalConfig.stageName, configOptions.stepName)

	//ToDo: Check for mandatory parameters

	myConfigJSON, _ := config.GetJSON(stepConfig.Config)

	fmt.Println(myConfigJSON)

	return nil
}

func addConfigFlags(cmd *cobra.Command) {

	//ToDo: support more output options, like https://kubernetes.io/docs/reference/kubectl/overview/#formatting-output
	cmd.Flags().StringVar(&configOptions.output, "output", "json", "Defines the output format")

	cmd.Flags().StringVar(&configOptions.parametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	cmd.Flags().StringVar(&configOptions.stepMetadata, "stepMetadata", "", "Step metadata, passed as path to yaml")
	cmd.Flags().StringVar(&configOptions.stepName, "stepName", "", "Name of the step for which configuration should be included")
	cmd.Flags().BoolVar(&configOptions.contextConfig, "contextConfig", false, "Defines if step context configuration should be loaded instead of step config")

	cmd.MarkFlagRequired("stepMetadata")
	cmd.MarkFlagRequired("stepName")

}

func openPiperFile(name string) (io.ReadCloser, error) {
	//ToDo: support also https as source
	if !strings.HasPrefix(name, "http") {
		return os.Open(name)
	}
	return nil, fmt.Errorf("file location not yet supported for '%v'", name)
}

func defaultsAndFilters(metadata *config.StepData) ([]io.ReadCloser, config.StepFilters, error) {
	if configOptions.contextConfig {
		defaults, err := metadata.GetContextDefaults(configOptions.stepName)
		if err != nil {
			return nil, config.StepFilters{}, errors.Wrap(err, "metadata: getting context defaults failed")
		}
		return []io.ReadCloser{defaults}, metadata.GetContextParameterFilters(), nil
	}
	//ToDo: retrieve default values from metadata
	return nil, metadata.GetParameterFilters(), nil
}
