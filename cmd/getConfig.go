package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/reporting"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"
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
			initStageName(false)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			err := generateConfig()
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				log.Entry().WithError(err).Fatal("failed to retrieve configuration")
			}
		},
	}

	addConfigFlags(createConfigCmd)
	return createConfigCmd
}

func generateConfig() error {

	var myConfig config.Config
	var stepConfig config.StepConfig

	metadata, err := resolveMetadata()
	if err != nil {
		return errors.Wrapf(err, "failed to resolve metadata: ")
	}

	fmt.Println(config.GetJSON(metadata))

	// prepare output resource directories:
	// this is needed in order to have proper directory permissions in case
	// resources written inside a container image with a different user
	// Remark: This is so far only relevant for Jenkins environments where getConfig is executed
	prepareOutputEnvironment(metadata.Spec.Outputs.Resources, GeneralConfig.EnvRootPath)

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
	cmd.Flags().StringVar(&configOptions.stepName, "stepName", "", "Step name, used to get step metadata if yaml path is not set")
	cmd.Flags().BoolVar(&configOptions.contextConfig, "contextConfig", false, "Defines if step context configuration should be loaded instead of step config")

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
	config.ApplyContainerConditions(metadata.Spec.Containers, stepConfig)

	//sidecars
	config.ApplyContainerConditions(metadata.Spec.Sidecars, stepConfig)

	//ToDo: remove all unnecessary sub maps?
	// e.g. extract delete() from applyContainerConditions - loop over all stepConfig.Config[param.Value] and remove ...
}

func prepareOutputEnvironment(outputResources []config.StepResources, envRootPath string) {
	for _, oResource := range outputResources {
		for _, oParam := range oResource.Parameters {
			paramPath := path.Join(envRootPath, oResource.Name, fmt.Sprint(oParam["name"]))
			if oParam["fields"] != nil {
				paramFields, ok := oParam["fields"].([]map[string]string)
				if ok && len(paramFields) > 0 {
					paramPath = path.Join(paramPath, paramFields[0]["name"])
				}
			}
			if _, err := os.Stat(filepath.Dir(paramPath)); os.IsNotExist(err) {
				log.Entry().Debugf("Creating directory: %v", filepath.Dir(paramPath))
				os.MkdirAll(filepath.Dir(paramPath), 0777)
			}
		}
	}

	// prepare additional output directories known to possibly create permission issues when created from within a container
	// ToDo: evaluate if we can rather call this only in the correct step context (we know the step when calling getConfig!)
	// Could this be part of the container definition in the step.yaml?
	stepOutputDirectories := []string{
		reporting.StepReportDirectory, // standard directory to collect md reports for pipelineCreateScanSummary
		ws.ReportsDirectory,           // standard directory for reports created by whitesourceExecuteScan
	}

	for _, dir := range stepOutputDirectories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Entry().Debugf("Creating directory: %v", dir)
			os.MkdirAll(dir, 0777)
		}
	}
}

func resolveMetadata() (config.StepData, error) {
	var metadata config.StepData
	if configOptions.stepMetadata != "" {
		metadataFile, err := configOptions.openFile(configOptions.stepMetadata)
		if err != nil {
			return metadata, errors.Wrap(err, "open failed")
		}

		err = metadata.ReadPipelineStepData(metadataFile)
		if err != nil {
			return metadata, errors.Wrap(err, "read failed")
		}
	} else {
		if configOptions.stepName != "" {
			metadataMap := GetAllStepMetadata()
			var ok bool
			metadata, ok = metadataMap[configOptions.stepName]
			if !ok {
				return metadata, errors.Errorf("could not retrieve by stepName %v", configOptions.stepName)
			}
		} else {
			return metadata, errors.Errorf("either one of stepMetadata or stepName parameter has to be passed")
		}
	}
	return metadata, nil
}
