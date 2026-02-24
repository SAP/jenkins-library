package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"errors"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"
	"github.com/spf13/cobra"
)

type ConfigCommandOptions struct {
	Output                        string // output format, so far only JSON, YAML
	OutputFile                    string // if set: path to file where the output should be written to
	ParametersJSON                string // parameters to be considered in JSON format
	StageConfig                   bool
	StageConfigAcceptedParameters []string
	StepMetadata                  string // metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	StepName                      string
	ContextConfig                 bool
	OpenFile                      func(s string, t map[string]string) (io.ReadCloser, error)
	SetVaultCredentials           bool
}

var configOptions ConfigCommandOptions

func SetConfigOptions(c ConfigCommandOptions) {
	configOptions.ContextConfig = c.ContextConfig
	configOptions.OpenFile = c.OpenFile
	configOptions.Output = c.Output
	configOptions.OutputFile = c.OutputFile
	configOptions.ParametersJSON = c.ParametersJSON
	configOptions.StageConfig = c.StageConfig
	configOptions.StageConfigAcceptedParameters = c.StageConfigAcceptedParameters
	configOptions.StepMetadata = c.StepMetadata
	configOptions.StepName = c.StepName
	configOptions.SetVaultCredentials = c.SetVaultCredentials
}

type getConfigUtils interface {
	FileExists(filename string) (bool, error)
	DirExists(path string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
}

type getConfigUtilsBundle struct {
	*piperutils.Files
}

func newGetConfigUtilsUtils() getConfigUtils {
	return &getConfigUtilsBundle{
		Files: &piperutils.Files{},
	}
}

// ConfigCommand is the entry command for loading the configuration of a pipeline step
func ConfigCommand() *cobra.Command {
	SetConfigOptions(ConfigCommandOptions{
		OpenFile: config.OpenPiperFile,
	})

	createConfigCmd := &cobra.Command{
		Use:   "getConfig",
		Short: "Loads the project 'Piper' configuration respecting defaults and parameters.",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
			initStageName(false)
			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			if err := generateConfigWrapper(cmd); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				log.Entry().WithError(err).Fatal("failed to retrieve configuration")
			}
		},
	}

	addConfigFlags(createConfigCmd)
	return createConfigCmd
}

// GetDockerImageValue provides Piper commands additional access to configuration of step execution image if required
func GetDockerImageValue(stepName string) (string, error) {
	configOptions.ContextConfig = true
	configOptions.StepName = stepName
	stepConfig, err := getConfig()
	if err != nil {
		return "", err
	}

	var dockerImageValue string
	dockerImageValue, ok := stepConfig.Config["dockerImage"].(string)
	if !ok {
		log.Entry().Infof("Config value of %v to compare with is not a string", stepConfig.Config["dockerImage"])
	}

	return dockerImageValue, nil
}

func getBuildToolFromStageConfig(stepName string) (string, error) {
	configOptions.ContextConfig = true
	configOptions.StepName = stepName
	stageConfig, err := GetStageConfig()
	if err != nil {
		return "", err
	}

	buildTool, ok := stageConfig.Config["buildTool"].(string)
	if !ok {
		log.Entry().Infof("Config value of %v to compare with is not a string", stageConfig.Config["buildTool"])
	}

	return buildTool, nil
}

// GetStageConfig provides Piper commands additional access to stage configuration if required.
// This allows steps to refer to configuration parameters which are not part of the step itself.
func GetStageConfig() (config.StepConfig, error) {
	myConfig := config.Config{}
	stepConfig := config.StepConfig{}
	projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)

	customConfig, err := configOptions.OpenFile(projectConfigFile, GeneralConfig.GitHubAccessTokens)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return stepConfig, fmt.Errorf("config: open configuration file '%v' failed: %w", projectConfigFile, err)
		}
		customConfig = nil
	}

	defaultConfig := []io.ReadCloser{}
	for _, f := range GeneralConfig.DefaultConfig {
		if configOptions.OpenFile == nil {
			return stepConfig, errors.New("config: open file function not set")
		}
		fc, err := configOptions.OpenFile(f, GeneralConfig.GitHubAccessTokens)
		// only create error for non-default values
		if err != nil && f != ".pipeline/defaults.yaml" {
			return stepConfig, fmt.Errorf("config: getting defaults failed: '%v': %w", f, err)
		}
		if err == nil {
			defaultConfig = append(defaultConfig, fc)
		}
	}

	return myConfig.GetStageConfig(GeneralConfig.ParametersJSON, customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, configOptions.StageConfigAcceptedParameters, GeneralConfig.StageName)
}

func getConfig() (config.StepConfig, error) {
	return getConfigWithFlagValues(nil)
}

func getConfigWithFlagValues(cmd *cobra.Command) (config.StepConfig, error) {
	var myConfig config.Config
	var stepConfig config.StepConfig
	var err error

	if configOptions.StageConfig {
		stepConfig, err = GetStageConfig()
		if err != nil {
			return stepConfig, fmt.Errorf("getting stage config failed: %w", err)
		}
		// add hooks (defaults + custom defaults) to stage-config.json output
		stepConfig.Config["hooks"] = stepConfig.HookConfig
	} else {
		log.Entry().Infof("Printing stepName %s", configOptions.StepName)
		if GeneralConfig.MetaDataResolver == nil {
			GeneralConfig.MetaDataResolver = GetAllStepMetadata
		}
		metadata, err := config.ResolveMetadata(GeneralConfig.GitHubAccessTokens, GeneralConfig.MetaDataResolver, configOptions.StepMetadata, configOptions.StepName)
		if err != nil {
			return stepConfig, fmt.Errorf("failed to resolve metadata: %w", err)
		}

		// prepare output resource directories:
		// this is needed in order to have proper directory permissions in case
		// resources written inside a container image with a different user
		// Remark: This is so far only relevant for Jenkins environments where getConfig is executed

		prepareOutputEnvironment(metadata.Spec.Outputs.Resources, GeneralConfig.EnvRootPath)

		envParams := metadata.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
		reportingEnvParams := config.ReportingParameters.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
		resourceParams := mergeResourceParameters(envParams, reportingEnvParams)

		projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)

		customConfig, err := configOptions.OpenFile(projectConfigFile, GeneralConfig.GitHubAccessTokens)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return stepConfig, fmt.Errorf("config: open configuration file '%v' failed: %w", projectConfigFile, err)
			}
			customConfig = nil
		}

		defaultConfig, paramFilter, err := defaultsAndFilters(&metadata, metadata.Metadata.Name)
		if err != nil {
			return stepConfig, fmt.Errorf("defaults: retrieving step defaults failed: %w", err)
		}

		for _, f := range GeneralConfig.DefaultConfig {
			fc, err := configOptions.OpenFile(f, GeneralConfig.GitHubAccessTokens)
			// only create error for non-default values
			if err != nil && f != ".pipeline/defaults.yaml" {
				return stepConfig, fmt.Errorf("config: getting defaults failed: '%v': %w", f, err)
			}
			if err == nil {
				defaultConfig = append(defaultConfig, fc)
			}
		}

		if configOptions.ContextConfig {
			metadata.Spec.Inputs.Parameters = []config.StepParameters{}
		}

		var flagValues map[string]interface{}
		if cmd != nil {
			flagValues = config.AvailableFlagValues(cmd, &paramFilter)
		}

		if configOptions.SetVaultCredentials {
			// add vault credentials so that configuration can be fetched from vault
			if GeneralConfig.VaultRoleID == "" {
				GeneralConfig.VaultRoleID = os.Getenv("PIPER_vaultAppRoleID")
			}
			if GeneralConfig.VaultRoleSecretID == "" {
				GeneralConfig.VaultRoleSecretID = os.Getenv("PIPER_vaultAppRoleSecretID")
			}
			if GeneralConfig.VaultToken == "" {
				GeneralConfig.VaultToken = os.Getenv("PIPER_vaultToken")
			}
			myConfig.SetVaultCredentials(GeneralConfig.VaultRoleID, GeneralConfig.VaultRoleSecretID, GeneralConfig.VaultToken)
		}

		stepConfig, err = myConfig.GetStepConfig(flagValues, GeneralConfig.ParametersJSON, customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, paramFilter, metadata, resourceParams, GeneralConfig.StageName, metadata.Metadata.Name)
		if err != nil {
			return stepConfig, fmt.Errorf("getting step config failed: %w", err)
		}

		// apply context conditions if context configuration is requested
		if configOptions.ContextConfig {
			applyContextConditions(metadata, &stepConfig)
		}
	}
	return stepConfig, nil
}

func generateConfigWrapper(cmd *cobra.Command) error {
	var formatter func(interface{}) (string, error)
	switch strings.ToLower(configOptions.Output) {
	case "yaml", "yml":
		formatter = config.GetYAML
	case "json":
		formatter = config.GetJSON
	default:
		formatter = config.GetJSON
	}
	return GenerateConfig(cmd, formatter)
}

func GenerateConfig(cmd *cobra.Command, formatter func(interface{}) (string, error)) error {
	utils := newGetConfigUtilsUtils()

	stepConfig, err := getConfigWithFlagValues(cmd)
	if err != nil {
		return err
	}

	myConfig, err := formatter(stepConfig.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if len(configOptions.OutputFile) > 0 {
		if err := utils.FileWrite(configOptions.OutputFile, []byte(myConfig), 0o666); err != nil {
			return fmt.Errorf("failed to write output file %v: %w", configOptions.OutputFile, err)
		}
		return nil
	}
	fmt.Println(myConfig)

	return nil
}

func addConfigFlags(cmd *cobra.Command) {
	// ToDo: support more output options, like https://kubernetes.io/docs/reference/kubectl/overview/#formatting-output
	cmd.Flags().StringVar(&configOptions.Output, "output", "json", "Defines the output format")
	cmd.Flags().StringVar(&configOptions.OutputFile, "outputFile", "", "Defines a file path. f set, the output will be written to the defines file")

	cmd.Flags().StringVar(&configOptions.ParametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	cmd.Flags().BoolVar(&configOptions.StageConfig, "stageConfig", false, "Defines if step stage configuration should be loaded and no step-specific config")
	cmd.Flags().StringArrayVar(&configOptions.StageConfigAcceptedParameters, "stageConfigAcceptedParams", []string{}, "Defines the parameters used for filtering stage/general configuration when accessing stage config")
	cmd.Flags().StringVar(&configOptions.StepMetadata, "stepMetadata", "", "Step metadata, passed as path to yaml")
	cmd.Flags().StringVar(&configOptions.StepName, "stepName", "", "Step name, used to get step metadata if yaml path is not set")
	cmd.Flags().BoolVar(&configOptions.ContextConfig, "contextConfig", false, "Defines if step context configuration should be loaded instead of step config")
	cmd.Flags().BoolVar(&configOptions.SetVaultCredentials, "setVaultCredentials", false, "Defines whether to set Vault credentials to enable fetching credentials from Vault or not")
}

func defaultsAndFilters(metadata *config.StepData, stepName string) ([]io.ReadCloser, config.StepFilters, error) {
	if configOptions.ContextConfig {
		defaults, err := metadata.GetContextDefaults(stepName)
		if err != nil {
			return nil, config.StepFilters{}, fmt.Errorf("metadata: getting context defaults failed: %w", err)
		}
		return []io.ReadCloser{defaults}, metadata.GetContextParameterFilters(), nil
	}
	// ToDo: retrieve default values from metadata
	return []io.ReadCloser{}, metadata.GetParameterFilters(), nil
}

func applyContextConditions(metadata config.StepData, stepConfig *config.StepConfig) {
	// consider conditions for context configuration

	// containers
	config.ApplyContainerConditions(metadata.Spec.Containers, stepConfig)

	// sidecars
	config.ApplyContainerConditions(metadata.Spec.Sidecars, stepConfig)

	// ToDo: remove all unnecessary sub maps?
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
			if _, err := os.Stat(filepath.Dir(paramPath)); errors.Is(err, os.ErrNotExist) {
				log.Entry().Debugf("Creating directory: %v", filepath.Dir(paramPath))
				_ = os.MkdirAll(filepath.Dir(paramPath), 0o777)
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
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			log.Entry().Debugf("Creating directory: %v", dir)
			_ = os.MkdirAll(dir, 0o777)
		}
	}
}
