package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
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
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			err := generateConfig()
			if err != nil {
				log.Entry().WithField("category", "config").WithError(err).Fatal("failed to retrieve configuration")
			}
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

	resourceParams := metadata.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")

	projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)

	customConfig, err := configOptions.openFile(projectConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "config: open configuration file '%v' failed", projectConfigFile)
		}
		customConfig = nil
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
		if err == nil {
			defaultConfig = append(defaultConfig, fc)
		}
	}

	var flags map[string]interface{}

	params := []config.StepParameters{}
	if !configOptions.contextConfig {
		params = metadata.Spec.Inputs.Parameters
	}

	stepConfig, err = myConfig.GetStepConfig(flags, GeneralConfig.ParametersJSON, customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, paramFilter, params, metadata.Spec.Inputs.Secrets, resourceParams, GeneralConfig.StageName, metadata.Metadata.Name, metadata.Metadata.Aliases)
	if err != nil {
		return errors.Wrap(err, "getting step config failed")
	}

	// apply context conditions if context configuration is requested
	if configOptions.contextConfig {
		applyContextConditions(metadata, &stepConfig)
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

	_ = cmd.MarkFlagRequired("stepMetadata")

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
	return []io.ReadCloser{}, metadata.GetParameterFilters(), nil
}

func applyContextConditions(metadata config.StepData, stepConfig *config.StepConfig) {
	//consider conditions for context configuration

	//containers
	applyContainerConditions(metadata.Spec.Containers, stepConfig)

	//sidecars
	applyContainerConditions(metadata.Spec.Sidecars, stepConfig)

	//ToDo: remove all unnecessary sub maps?
	// e.g. extract delete() from applyContainerConditions - loop over all stepConfig.Config[param.Value] and remove ...
}

func applyContainerConditions(containers []config.Container, stepConfig *config.StepConfig) {
	for _, container := range containers {
		if len(container.Conditions) > 0 {
			for _, param := range container.Conditions[0].Params {
				if container.Conditions[0].ConditionRef == "strings-equal" && stepConfig.Config[param.Name] == param.Value {
					var containerConf map[string]interface{}
					if stepConfig.Config[param.Value] != nil {
						containerConf = stepConfig.Config[param.Value].(map[string]interface{})
						for key, value := range containerConf {
							if stepConfig.Config[key] == nil {
								stepConfig.Config[key] = value
							}
						}
						delete(stepConfig.Config, param.Value)
					}
				}
			}
		}
	}
}
