package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// GeneralConfigOptions contains all global configuration options for piper binary
type GeneralConfigOptions struct {
	CorrelationID        string
	CustomConfig         string
	DefaultConfig        []string //ordered list of Piper default configurations. Can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	IgnoreCustomDefaults bool
	ParametersJSON       string
	EnvRootPath          string
	NoTelemetry          bool
	StageName            string
	StepConfigJSON       string
	StepMetadata         string //metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	StepName             string
	Verbose              bool
	LogFormat            string
	VaultRoleID          string
	VaultRoleSecretID    string
	HookConfig           HookConfiguration
}

// HookConfiguration contains the configuration for supported hooks, so far only Sentry is supported.
type HookConfiguration struct {
	SentryConfig SentryConfiguration `json:"sentry,omitempty"`
}

// SentryConfiguration defines the configuration options for the Sentry logging system
type SentryConfiguration struct {
	Dsn string `json:"dsn,omitempty"`
}

var rootCmd = &cobra.Command{
	Use:   "piper",
	Short: "Executes CI/CD steps from project 'Piper' ",
	Long: `
This project 'Piper' binary provides a CI/CD step library.
It contains many steps which can be used within CI/CD systems as well as directly on e.g. a developer's machine.
`,
}

// GeneralConfig contains global configuration flags for piper binary
var GeneralConfig GeneralConfigOptions

// Execute is the starting point of the piper command line tool
func Execute() {

	rootCmd.AddCommand(ArtifactPrepareVersionCommand())
	rootCmd.AddCommand(ConfigCommand())
	rootCmd.AddCommand(ContainerSaveImageCommand())
	rootCmd.AddCommand(CommandLineCompletionCommand())
	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(DetectExecuteScanCommand())
	rootCmd.AddCommand(KarmaExecuteTestsCommand())
	rootCmd.AddCommand(SonarExecuteScanCommand())
	rootCmd.AddCommand(KubernetesDeployCommand())
	rootCmd.AddCommand(XsDeployCommand())
	rootCmd.AddCommand(GithubCheckBranchProtectionCommand())
	rootCmd.AddCommand(GithubCreatePullRequestCommand())
	rootCmd.AddCommand(GithubPublishReleaseCommand())
	rootCmd.AddCommand(GithubSetCommitStatusCommand())
	rootCmd.AddCommand(GitopsUpdateDeploymentCommand())
	rootCmd.AddCommand(CloudFoundryDeleteServiceCommand())
	rootCmd.AddCommand(AbapEnvironmentPullGitRepoCommand())
	rootCmd.AddCommand(AbapEnvironmentCloneGitRepoCommand())
	rootCmd.AddCommand(AbapEnvironmentCheckoutBranchCommand())
	rootCmd.AddCommand(AbapEnvironmentCreateSystemCommand())
	rootCmd.AddCommand(CheckmarxExecuteScanCommand())
	rootCmd.AddCommand(FortifyExecuteScanCommand())
	rootCmd.AddCommand(MtaBuildCommand())
	rootCmd.AddCommand(ProtecodeExecuteScanCommand())
	rootCmd.AddCommand(MavenExecuteCommand())
	rootCmd.AddCommand(CloudFoundryCreateServiceKeyCommand())
	rootCmd.AddCommand(MavenBuildCommand())
	rootCmd.AddCommand(MavenExecuteIntegrationCommand())
	rootCmd.AddCommand(MavenExecuteStaticCodeChecksCommand())
	rootCmd.AddCommand(NexusUploadCommand())
	rootCmd.AddCommand(AbapEnvironmentRunATCCheckCommand())
	rootCmd.AddCommand(NpmExecuteScriptsCommand())
	rootCmd.AddCommand(NpmExecuteLintCommand())
	rootCmd.AddCommand(GctsCreateRepositoryCommand())
	rootCmd.AddCommand(GctsExecuteABAPUnitTestsCommand())
	rootCmd.AddCommand(GctsDeployCommand())
	rootCmd.AddCommand(MalwareExecuteScanCommand())
	rootCmd.AddCommand(CloudFoundryCreateServiceCommand())
	rootCmd.AddCommand(CloudFoundryDeployCommand())
	rootCmd.AddCommand(GctsRollbackCommand())
	rootCmd.AddCommand(WhitesourceExecuteScanCommand())
	rootCmd.AddCommand(GctsCloneRepositoryCommand())
	rootCmd.AddCommand(JsonApplyPatchCommand())
	rootCmd.AddCommand(KanikoExecuteCommand())
	rootCmd.AddCommand(AbapEnvironmentAssemblePackagesCommand())
	rootCmd.AddCommand(AbapAddonAssemblyKitCheckCVsCommand())
	rootCmd.AddCommand(AbapAddonAssemblyKitCheckPVCommand())
	rootCmd.AddCommand(AbapAddonAssemblyKitCreateTargetVectorCommand())
	rootCmd.AddCommand(AbapAddonAssemblyKitPublishTargetVectorCommand())
	rootCmd.AddCommand(AbapAddonAssemblyKitRegisterPackagesCommand())
	rootCmd.AddCommand(AbapAddonAssemblyKitReleasePackagesCommand())
	rootCmd.AddCommand(AbapAddonAssemblyKitReserveNextPackagesCommand())
	rootCmd.AddCommand(CloudFoundryCreateSpaceCommand())
	rootCmd.AddCommand(CloudFoundryDeleteSpaceCommand())

	addRootFlags(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		log.Entry().WithError(err).Fatal("configuration error")
	}
}

func addRootFlags(rootCmd *cobra.Command) {

	rootCmd.PersistentFlags().StringVar(&GeneralConfig.CorrelationID, "correlationID", os.Getenv("PIPER_correlationID"), "ID for unique identification of a pipeline run")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.CustomConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&GeneralConfig.DefaultConfig, "defaultConfig", []string{".pipeline/defaults.yaml"}, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().BoolVar(&GeneralConfig.IgnoreCustomDefaults, "ignoreCustomDefaults", false, "Disables evaluation of the parameter 'customDefaults' in the pipeline configuration file")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.ParametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.EnvRootPath, "envRootPath", ".pipeline", "Root path to Piper pipeline shared environments")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StageName, "stageName", "", "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVar(&GeneralConfig.NoTelemetry, "noTelemetry", false, "Disables telemetry reporting")
	rootCmd.PersistentFlags().BoolVarP(&GeneralConfig.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.LogFormat, "logFormat", "default", "Log format to use. Options: default, timestamp, plain, full.")

}

const stageNameEnvKey = "STAGE_NAME"

// initStageName initializes GeneralConfig.StageName from either GeneralConfig.ParametersJSON
// or the environment variable 'STAGE_NAME', unless it has been provided as command line option.
// Log output needs to be suppressed via outputToLog by the getConfig step.
func initStageName(outputToLog bool) {
	var stageNameSource string
	if outputToLog {
		defer func() {
			log.Entry().Infof("Using stageName '%s' from %s", GeneralConfig.StageName, stageNameSource)
		}()
	}

	if GeneralConfig.StageName != "" {
		// Means it was given as command line argument and has the highest precedence
		stageNameSource = "command line arguments"
		return
	}

	// Use stageName from ENV as fall-back, for when extracting it from parametersJSON fails below
	GeneralConfig.StageName = os.Getenv(stageNameEnvKey)
	stageNameSource = fmt.Sprintf("env variable '%s'", stageNameEnvKey)

	if len(GeneralConfig.ParametersJSON) == 0 {
		return
	}

	var params map[string]interface{}
	err := json.Unmarshal([]byte(GeneralConfig.ParametersJSON), &params)
	if err != nil {
		if outputToLog {
			log.Entry().Infof("Failed to extract 'stageName' from parametersJSON: %v", err)
		}
		return
	}

	stageName, hasKey := params["stageName"]
	if !hasKey {
		return
	}

	if stageNameString, ok := stageName.(string); ok && stageNameString != "" {
		stageNameSource = "parametersJSON"
		GeneralConfig.StageName = stageNameString
	}
}

// PrepareConfig reads step configuration from various sources and merges it (defaults, config file, flags, ...)
func PrepareConfig(cmd *cobra.Command, metadata *config.StepData, stepName string, options interface{}, openFile func(s string) (io.ReadCloser, error)) error {

	log.SetFormatter(GeneralConfig.LogFormat)

	initStageName(true)

	filters := metadata.GetParameterFilters()

	// add telemetry parameter "collectTelemetryData" to ALL, GENERAL and PARAMETER filters
	filters.All = append(filters.All, "collectTelemetryData")
	filters.General = append(filters.General, "collectTelemetryData")
	filters.Parameters = append(filters.Parameters, "collectTelemetryData")

	resourceParams := metadata.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	flagValues := config.AvailableFlagValues(cmd, &filters)

	var myConfig config.Config
	var stepConfig config.StepConfig

	// add vault credentials so that configuration can be fetched from vault
	if GeneralConfig.VaultRoleID == "" {
		GeneralConfig.VaultRoleID = os.Getenv("PIPER_vaultAppRoleID")
	}
	if GeneralConfig.VaultRoleSecretID == "" {
		GeneralConfig.VaultRoleSecretID = os.Getenv("PIPER_vaultAppRoleSecretID")
	}
	myConfig.SetVaultCredentials(GeneralConfig.VaultRoleID, GeneralConfig.VaultRoleSecretID)

	if len(GeneralConfig.StepConfigJSON) != 0 {
		// ignore config & defaults in favor of passed stepConfigJSON
		stepConfig = config.GetStepConfigWithJSON(flagValues, GeneralConfig.StepConfigJSON, filters)
		log.Entry().Infof("Project config: passed via JSON")
		log.Entry().Infof("Project defaults: passed via JSON")
	} else {
		// use config & defaults
		var customConfig io.ReadCloser
		var err error
		//accept that config file and defaults cannot be loaded since both are not mandatory here
		{
			projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)
			if exists, err := piperutils.FileExists(projectConfigFile); exists {
				log.Entry().Infof("Project config: '%s'", projectConfigFile)
				if customConfig, err = openFile(projectConfigFile); err != nil {
					return errors.Wrapf(err, "Cannot read '%s'", projectConfigFile)
				}
			} else {
				log.Entry().Infof("Project config: NONE ('%s' does not exist)", projectConfigFile)
				customConfig = nil
			}
		}
		var defaultConfig []io.ReadCloser
		if len(GeneralConfig.DefaultConfig) == 0 {
			log.Entry().Info("Project defaults: NONE")
		}
		for _, projectDefaultFile := range GeneralConfig.DefaultConfig {
			fc, err := openFile(projectDefaultFile)
			// only create error for non-default values
			if err != nil {
				if projectDefaultFile != ".pipeline/defaults.yaml" {
					log.Entry().Infof("Project defaults: '%s'", projectDefaultFile)
					return errors.Wrapf(err, "Cannot read '%s'", projectDefaultFile)
				}
			} else {
				log.Entry().Infof("Project defaults: '%s'", projectDefaultFile)
				defaultConfig = append(defaultConfig, fc)
			}
		}
		stepConfig, err = myConfig.GetStepConfig(flagValues, GeneralConfig.ParametersJSON, customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, filters, metadata.Spec.Inputs.Parameters, metadata.Spec.Inputs.Secrets, resourceParams, GeneralConfig.StageName, stepName, metadata.Metadata.Aliases)
		if err != nil {
			return errors.Wrap(err, "retrieving step configuration failed")
		}
	}

	if fmt.Sprintf("%v", stepConfig.Config["collectTelemetryData"]) == "false" {
		GeneralConfig.NoTelemetry = true
	}

	if !GeneralConfig.Verbose && stepConfig.Config["verbose"] != nil {
		if verboseValue, ok := stepConfig.Config["verbose"].(bool); ok {
			log.SetVerbose(verboseValue)
		} else {
			return fmt.Errorf("invalid value for parameter verbose: '%v'", stepConfig.Config["verbose"])
		}
	}

	stepConfig.Config = checkTypes(stepConfig.Config, options)
	confJSON, _ := json.Marshal(stepConfig.Config)
	_ = json.Unmarshal(confJSON, &options)

	config.MarkFlagsWithValue(cmd, stepConfig)

	retrieveHookConfig(stepConfig.HookConfig, &GeneralConfig.HookConfig)

	return nil
}

func retrieveHookConfig(source *json.RawMessage, target *HookConfiguration) {
	if source != nil {
		log.Entry().Info("Retrieving hook configuration")
		err := json.Unmarshal(*source, target)
		if err != nil {
			log.Entry().Warningf("Failed to retrieve hook configuration: %v", err)
		}
	}
}

var errIncompatibleTypes = fmt.Errorf("incompatible types")

func checkTypes(config map[string]interface{}, options interface{}) map[string]interface{} {
	optionsType := getStepOptionsStructType(options)

	for paramName := range config {
		optionsField := findStructFieldByJSONTag(paramName, optionsType)
		if optionsField == nil {
			continue
		}

		if config[paramName] == nil {
			// There is a key, but no value. This can result from merging values from the CPE.
			continue
		}

		paramValueType := reflect.ValueOf(config[paramName])
		if optionsField.Type.Kind() == paramValueType.Kind() {
			// Types already match, nothing to do
			continue
		}

		var typeError error = nil

		switch paramValueType.Kind() {
		case reflect.String:
			typeError = convertValueFromString(config, optionsField, paramName, paramValueType.String())
		case reflect.Float32, reflect.Float64:
			typeError = convertValueFromFloat(config, optionsField, paramName, paramValueType.Float())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			typeError = convertValueFromInt(config, optionsField, paramName, paramValueType.Int())
		default:
			log.Entry().Warnf("Config value for '%s' is of unexpected type %s, expected %s. "+
				"The value may be ignored as a result. To avoid any risk, specify this value with explicit type.",
				paramName, paramValueType.Kind(), optionsField.Type.Kind())
		}

		if typeError != nil {
			typeError = fmt.Errorf("config value for '%s' is of unexpected type %s, expected %s: %w",
				paramName, paramValueType.Kind(), optionsField.Type.Kind(), typeError)
			log.SetErrorCategory(log.ErrorConfiguration)
			log.Entry().WithError(typeError).Fatal("type error in configuration")
		}
	}
	return config
}

func convertValueFromString(config map[string]interface{}, optionsField *reflect.StructField, paramName, paramValue string) error {
	switch optionsField.Type.Kind() {
	case reflect.Slice, reflect.Array:
		// Could do automatic conversion for those types in theory,
		// but that might obscure what really happens in error cases.
		return fmt.Errorf("expected type to be a list (or slice, or array) but got string")
	case reflect.Bool:
		// Sensible to convert strings "true"/"false" to respective boolean values as it is
		// common practice to write booleans as string in yaml files.
		paramValue = strings.ToLower(paramValue)
		if paramValue == "true" {
			config[paramName] = true
			return nil
		} else if paramValue == "false" {
			config[paramName] = false
			return nil
		}
	}

	return errIncompatibleTypes
}

func convertValueFromFloat(config map[string]interface{}, optionsField *reflect.StructField, paramName string, paramValue float64) error {
	switch optionsField.Type.Kind() {
	case reflect.String:
		val := strconv.FormatFloat(paramValue, 'f', -1, 64)
		// if Sprinted value and val are equal, we can be pretty sure that the result fits
		// for very large numbers for example an exponential format is printed
		if val == fmt.Sprint(paramValue) {
			config[paramName] = val
			return nil
		}
		// allow float numbers containing a decimal separator
		if strings.Contains(val, ".") {
			config[paramName] = val
			return nil
		}
		// if now no decimal separator is available we cannot be sure that the result is correct:
		// long numbers like e.g. 73554900100200011600 will not be represented correctly after reading the yaml
		// thus we cannot assume that the string is correct.
		// short numbers will be handled as int anyway
		return errIncompatibleTypes
	case reflect.Float32:
		config[paramName] = float32(paramValue)
		return nil
	case reflect.Float64:
		config[paramName] = paramValue
		return nil
	case reflect.Int:
		// Treat as type-mismatch only in case the conversion would be lossy.
		// In that case, the json.Unmarshall() would indeed just drop it, so we want to fail.
		if float64(int(paramValue)) == paramValue {
			config[paramName] = int(paramValue)
			return nil
		}
	}

	return errIncompatibleTypes
}

func convertValueFromInt(config map[string]interface{}, optionsField *reflect.StructField, paramName string, paramValue int64) error {
	switch optionsField.Type.Kind() {
	case reflect.String:
		config[paramName] = strconv.FormatInt(paramValue, 10)
		return nil
	case reflect.Float32:
		config[paramName] = float32(paramValue)
		return nil
	case reflect.Float64:
		config[paramName] = float64(paramValue)
		return nil
	}

	return errIncompatibleTypes
}

func findStructFieldByJSONTag(tagName string, optionsType reflect.Type) *reflect.StructField {
	for i := 0; i < optionsType.NumField(); i++ {
		field := optionsType.Field(i)
		tag := field.Tag.Get("json")
		if tagName == tag || tagName+",omitempty" == tag {
			return &field
		}
	}
	return nil
}

func getStepOptionsStructType(stepOptions interface{}) reflect.Type {
	typedOptions := reflect.ValueOf(stepOptions)
	if typedOptions.Kind() == reflect.Ptr {
		typedOptions = typedOptions.Elem()
	}
	return typedOptions.Type()
}

func getProjectConfigFile(name string) string {

	var altName string
	if ext := filepath.Ext(name); ext == ".yml" {
		altName = fmt.Sprintf("%v.yaml", strings.TrimSuffix(name, ext))
	} else if ext == "yaml" {
		altName = fmt.Sprintf("%v.yml", strings.TrimSuffix(name, ext))
	}

	fileExists, _ := piperutils.FileExists(name)
	altExists, _ := piperutils.FileExists(altName)

	// configured filename will always take precedence, even if not existing
	if !fileExists && altExists {
		return altName
	}
	return name
}
