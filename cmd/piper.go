package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// GeneralConfigOptions contains all global configuration options for piper binary
type GeneralConfigOptions struct {
	CorrelationID  string
	CustomConfig   string
	DefaultConfig  []string //ordered list of Piper default configurations. Can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	ParametersJSON string
	EnvRootPath    string
	NoTelemetry    bool
	StageName      string
	StepConfigJSON string
	StepMetadata   string //metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	StepName       string
	Verbose        bool
	LogFormat      string
	HookConfig     HookConfiguration
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
	//ToDo: respect stageName to also come from parametersJSON -> first env.STAGE_NAME, second: parametersJSON, third: flag
}

// GeneralConfig contains global configuration flags for piper binary
var GeneralConfig GeneralConfigOptions

// Execute is the starting point of the piper command line tool
func Execute() {

	rootCmd.AddCommand(ArtifactPrepareVersionCommand())
	rootCmd.AddCommand(ConfigCommand())
	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(DetectExecuteScanCommand())
	rootCmd.AddCommand(KarmaExecuteTestsCommand())
	rootCmd.AddCommand(SonarExecuteScanCommand())
	rootCmd.AddCommand(KubernetesDeployCommand())
	rootCmd.AddCommand(XsDeployCommand())
	rootCmd.AddCommand(GithubPublishReleaseCommand())
	rootCmd.AddCommand(GithubCreatePullRequestCommand())
	rootCmd.AddCommand(CloudFoundryDeleteServiceCommand())
	rootCmd.AddCommand(AbapEnvironmentPullGitRepoCommand())
	rootCmd.AddCommand(CheckmarxExecuteScanCommand())
	rootCmd.AddCommand(MtaBuildCommand())
	rootCmd.AddCommand(ProtecodeExecuteScanCommand())
	rootCmd.AddCommand(MavenExecuteCommand())
	rootCmd.AddCommand(CloudFoundryCreateServiceKeyCommand())
	rootCmd.AddCommand(MavenBuildCommand())
	rootCmd.AddCommand(MavenExecuteStaticCodeChecksCommand())
	rootCmd.AddCommand(NexusUploadCommand())
	rootCmd.AddCommand(NpmExecuteScriptsCommand())
	rootCmd.AddCommand(GctsCreateRepositoryCommand())
	rootCmd.AddCommand(MalwareExecuteScanCommand())

	addRootFlags(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		// in case we end up here we know that something in the PreRunE function went wrong
		// and thus this indicates a configuration issue
		log.Entry().WithError(err).WithField("category", "configuration").Fatal("configuration error")
	}
}

func addRootFlags(rootCmd *cobra.Command) {

	rootCmd.PersistentFlags().StringVar(&GeneralConfig.CorrelationID, "correlationID", os.Getenv("PIPER_correlationID"), "ID for unique identification of a pipeline run")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.CustomConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&GeneralConfig.DefaultConfig, "defaultConfig", []string{".pipeline/defaults.yaml"}, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.ParametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.EnvRootPath, "envRootPath", ".pipeline", "Root path to Piper pipeline shared environments")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StageName, "stageName", os.Getenv("STAGE_NAME"), "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVar(&GeneralConfig.NoTelemetry, "noTelemetry", false, "Disables telemetry reporting")
	rootCmd.PersistentFlags().BoolVarP(&GeneralConfig.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.LogFormat, "logFormat", "default", "Log format to use. Options: default, timestamp, plain, full.")

}

// PrepareConfig reads step configuration from various sources and merges it (defaults, config file, flags, ...)
func PrepareConfig(cmd *cobra.Command, metadata *config.StepData, stepName string, options interface{}, openFile func(s string) (io.ReadCloser, error)) error {

	filters := metadata.GetParameterFilters()

	// add telemetry parameter "collectTelemetryData" to ALL, GENERAL and PARAMETER filters
	filters.All = append(filters.All, "collectTelemetryData")
	filters.General = append(filters.General, "collectTelemetryData")
	filters.Parameters = append(filters.Parameters, "collectTelemetryData")

	resourceParams := metadata.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")

	flagValues := config.AvailableFlagValues(cmd, &filters)

	var myConfig config.Config
	var stepConfig config.StepConfig

	log.SetFormatter(GeneralConfig.LogFormat)

	if len(GeneralConfig.StepConfigJSON) != 0 {
		// ignore config & defaults in favor of passed stepConfigJSON
		stepConfig = config.GetStepConfigWithJSON(flagValues, GeneralConfig.StepConfigJSON, filters)
	} else {
		// use config & defaults
		var customConfig io.ReadCloser
		var err error
		//accept that config file and defaults cannot be loaded since both are not mandatory here
		{
			projectConfigFile := getProjectConfigFile(GeneralConfig.CustomConfig)

			exists, err := piperutils.FileExists(projectConfigFile)
			if exists {
				if customConfig, err = openFile(projectConfigFile); err != nil {
					return errors.Wrapf(err, "Cannot read '%s'", projectConfigFile)
				}
			} else {
				log.Entry().Infof("Project config file '%s' does not exist. No project configuration available.", projectConfigFile)
				customConfig = nil
			}

		}
		var defaultConfig []io.ReadCloser
		for _, f := range GeneralConfig.DefaultConfig {
			fc, err := openFile(f)
			// only create error for non-default values
			if err != nil && f != ".pipeline/defaults.yaml" {
				return errors.Wrapf(err, "config: getting defaults failed: '%v'", f)
			}
			if err == nil {
				defaultConfig = append(defaultConfig, fc)
				log.Entry().Infof("Added default config '%s'", f)
			}
		}

		stepConfig, err = myConfig.GetStepConfig(flagValues, GeneralConfig.ParametersJSON, customConfig, defaultConfig, filters, metadata.Spec.Inputs.Parameters, metadata.Spec.Inputs.Secrets, resourceParams, GeneralConfig.StageName, stepName, metadata.Metadata.Aliases)
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

	for name, v := range stepConfig.HookConfig {
		if name == "sentry" {
			hookConfig, _ := v.MarshalJSON()
			_ = json.Unmarshal(hookConfig, &GeneralConfig.HookConfig.SentryConfig)
		}
	}

	return nil
}

func checkTypes(config map[string]interface{}, options interface{}) map[string]interface{} {
	optionsType := getStepOptionsStructType(options)

	for paramName := range config {
		optionsField := findStructFieldByJSONTag(paramName, optionsType)
		if optionsField == nil {
			continue
		}

		paramValueType := reflect.ValueOf(config[paramName])
		if paramValueType.Kind() != reflect.String {
			// Type check is limited to strings at the moment
			continue
		}

		paramValue := paramValueType.String()
		logWarning := true

		switch optionsField.Type.Kind() {
		case reflect.String:
			// Types match, ignore
			logWarning = false
		case reflect.Slice, reflect.Array:
			// Could do automatic conversion for those types in theory,
			// but that might obscure what really happens in error cases.
			log.Entry().Fatalf("Type mismatch in configuration for option '%s'. Expected type to be a list (or slice, or array) but got %s.", paramName, paramValueType.Kind())
		case reflect.Bool:
			// Sensible to convert strings "true"/"false" to respective boolean values as it is
			// common practice to write booleans as string in yaml files.
			paramValue = strings.ToLower(paramValue)
			if paramValue == "true" {
				config[paramName] = true
				logWarning = false
			} else if paramValue == "false" {
				config[paramName] = false
				logWarning = false
			}
		}

		if logWarning {
			log.Entry().Warnf("Config value for '%s' is of unexpected type and is ignored", paramName)
		}
	}
	return config
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
