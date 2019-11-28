package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type configCommandOptions struct {
	output         string //output format, so far only JSON
	parametersJSON string //parameters to be considered in JSON format
	stepMetadata   string //metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	stepName       string
	contextConfig  bool
	openFile       func(s string) (io.ReadCloser, error)
}

var configOptions configCommandOptions

// ConfigCommand is the entry command for loading the configuration of a pipeline step
func ConfigCommand() *cobra.Command {

	configOptions.openFile = config.OpenPiperFile
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
	var stepConfig config.StepConfig

	var metadata config.StepData
	metadataFile, err := configOptions.openFile(configOptions.stepMetadata)
	if err != nil {
		return errors.Wrap(err, "metadata: open failed")
	}

	err = metadata.ReadPipelineStepData(metadataFile)
	if err != nil {
		return errors.Wrap(err, "metadata: read failed")
	}

	var customConfig io.ReadCloser
	if exists, e := piperutils.FileExists(GeneralConfig.CustomConfig); e == nil {
		if exists {
			customConfig, err = configOptions.openFile(GeneralConfig.CustomConfig)
			if err != nil {
				return errors.Wrap(err, "config: open failed")
			}
		}
	} else {
		return e
	}

	defaultConfig, paramFilter, err := defaultsAndFilters(&metadata, metadata.Metadata.Name)
	if err != nil {
		return errors.Wrap(err, "defaults: retrieving step defaults failed")
	}

	for _, f := range GeneralConfig.DefaultConfig {
		fc, err := configOptions.openFile(f)
		// only create error for non-default values
		if err != nil && f != ".pipeline/defaults.yaml" {
			return errors.Wrapf(err, "config: getting defaults failed: '%v'", f)
		}
		defaultConfig = append(defaultConfig, fc)
	}

	var flags map[string]interface{}

	params := []config.StepParameters{}
	if !configOptions.contextConfig {
		params = metadata.Spec.Inputs.Parameters
	}

	stepConfig, err = myConfig.GetStepConfig(flags, GeneralConfig.ParametersJSON, customConfig, defaultConfig, paramFilter, params, GeneralConfig.StageName, metadata.Metadata.Name)
	if err != nil {
		return errors.Wrap(err, "getting step config failed")
	}

	myConfigJSON, _ := config.GetJSON(stepConfig.Config)

	fmt.Println(myConfigJSON)

	return nil
}

func addConfigFlags(cmd *cobra.Command) {

	//ToDo: support more output options, like https://kubernetes.io/docs/reference/kubectl/overview/#formatting-output
	cmd.Flags().StringVar(&configOptions.output, "output", "json", "Defines the output format")

	cmd.Flags().StringVar(&configOptions.parametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	cmd.Flags().StringVar(&configOptions.stepMetadata, "stepMetadata", "", "Step metadata, passed as path to yaml")
	cmd.Flags().BoolVar(&configOptions.contextConfig, "contextConfig", false, "Defines if step context configuration should be loaded instead of step config")

	cmd.MarkFlagRequired("stepMetadata")

}

func defaultsAndFilters(metadata *config.StepData, stepName string) ([]io.ReadCloser, config.StepFilters, error) {
	if configOptions.contextConfig {
		defaults, err := metadata.GetContextDefaults(stepName)
		if err != nil {
			return nil, config.StepFilters{}, errors.Wrap(err, "metadata: getting context defaults failed")
		}
		return []io.ReadCloser{defaults}, metadata.GetContextParameterFilters(), nil
	}
	//ToDo: retrieve default values from metadata
	return nil, metadata.GetParameterFilters(), nil
}
