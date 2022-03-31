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

	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// GeneralConfigOptions contains all global configuration options for piper binary
type GeneralConfigOptions struct {
	GitHubAccessTokens   map[string]string // map of tokens with url as key in order to maintain url-specific tokens
	CorrelationID        string
	CustomConfig         string
	GitHubTokens         []string // list of entries in form of <server>:<token> to allow token authentication for downloading config / defaults
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
	VaultToken           string
	VaultServerURL       string
	VaultNamespace       string
	VaultPath            string
	HookConfig           HookConfiguration
	MetaDataResolver     func() map[string]config.StepData
	GCPJsonKeyFilePath   string
	GCSFolderPath        string
	GCSBucketId          string
	GCSSubFolder         string
}

// HookConfiguration contains the configuration for supported hooks, so far ANS, Sentry and Splunk are supported.
type HookConfiguration struct {
	SentryConfig SentryConfiguration `json:"sentry,omitempty"`
	SplunkConfig SplunkConfiguration `json:"splunk,omitempty"`
	ANSConfig    ans.Configuration   `json:"ans,omitempty"`
}

// SentryConfiguration defines the configuration options for the Sentry logging system
type SentryConfiguration struct {
	Dsn string `json:"dsn,omitempty"`
}

// SplunkConfiguration defines the configuration options for the Splunk logging system
type SplunkConfiguration struct {
	Dsn      string `json:"dsn,omitempty"`
	Token    string `json:"token,omitempty"`
	Index    string `json:"index,omitempty"`
	SendLogs bool   `json:"sendLogs"`
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
	rootCmd.AddCommand(DefaultsCommand())
	rootCmd.AddCommand(ContainerSaveImageCommand())
	rootCmd.AddCommand(CommandLineCompletionCommand())
	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(DetectExecuteScanCommand())
	rootCmd.AddCommand(HadolintExecuteCommand())
	rootCmd.AddCommand(KarmaExecuteTestsCommand())
	rootCmd.AddCommand(UiVeri5ExecuteTestsCommand())
	rootCmd.AddCommand(SonarExecuteScanCommand())
	rootCmd.AddCommand(KubernetesDeployCommand())
	rootCmd.AddCommand(HelmExecuteCommand())
	rootCmd.AddCommand(XsDeployCommand())
	rootCmd.AddCommand(GithubCheckBranchProtectionCommand())
	rootCmd.AddCommand(GithubCommentIssueCommand())
	rootCmd.AddCommand(GithubCreateIssueCommand())
	rootCmd.AddCommand(GithubCreatePullRequestCommand())
	rootCmd.AddCommand(GithubPublishReleaseCommand())
	rootCmd.AddCommand(GithubSetCommitStatusCommand())
	rootCmd.AddCommand(GitopsUpdateDeploymentCommand())
	rootCmd.AddCommand(CloudFoundryDeleteServiceCommand())
	rootCmd.AddCommand(AbapEnvironmentPullGitRepoCommand())
	rootCmd.AddCommand(AbapEnvironmentCloneGitRepoCommand())
	rootCmd.AddCommand(AbapEnvironmentCheckoutBranchCommand())
	rootCmd.AddCommand(AbapEnvironmentCreateTagCommand())
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
	rootCmd.AddCommand(AbapEnvironmentPushATCSystemConfigCommand())
	rootCmd.AddCommand(AbapEnvironmentRunATCCheckCommand())
	rootCmd.AddCommand(NpmExecuteScriptsCommand())
	rootCmd.AddCommand(NpmExecuteLintCommand())
	rootCmd.AddCommand(GctsCreateRepositoryCommand())
	rootCmd.AddCommand(GctsExecuteABAPQualityChecksCommand())
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
	rootCmd.AddCommand(CnbBuildCommand())
	rootCmd.AddCommand(AbapEnvironmentBuildCommand())
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
	rootCmd.AddCommand(VaultRotateSecretIdCommand())
	rootCmd.AddCommand(IsChangeInDevelopmentCommand())
	rootCmd.AddCommand(TransportRequestUploadCTSCommand())
	rootCmd.AddCommand(TransportRequestUploadRFCCommand())
	rootCmd.AddCommand(NewmanExecuteCommand())
	rootCmd.AddCommand(IntegrationArtifactDeployCommand())
	rootCmd.AddCommand(TransportRequestUploadSOLMANCommand())
	rootCmd.AddCommand(IntegrationArtifactUpdateConfigurationCommand())
	rootCmd.AddCommand(IntegrationArtifactGetMplStatusCommand())
	rootCmd.AddCommand(IntegrationArtifactGetServiceEndpointCommand())
	rootCmd.AddCommand(IntegrationArtifactDownloadCommand())
	rootCmd.AddCommand(AbapEnvironmentAssembleConfirmCommand())
	rootCmd.AddCommand(IntegrationArtifactUploadCommand())
	rootCmd.AddCommand(IntegrationArtifactTriggerIntegrationTestCommand())
	rootCmd.AddCommand(IntegrationArtifactUnDeployCommand())
	rootCmd.AddCommand(IntegrationArtifactResourceCommand())
	rootCmd.AddCommand(TerraformExecuteCommand())
	rootCmd.AddCommand(ContainerExecuteStructureTestsCommand())
	rootCmd.AddCommand(GaugeExecuteTestsCommand())
	rootCmd.AddCommand(BatsExecuteTestsCommand())
	rootCmd.AddCommand(PipelineCreateScanSummaryCommand())
	rootCmd.AddCommand(TransportRequestDocIDFromGitCommand())
	rootCmd.AddCommand(TransportRequestReqIDFromGitCommand())
	rootCmd.AddCommand(WritePipelineEnv())
	rootCmd.AddCommand(ReadPipelineEnv())
	rootCmd.AddCommand(InfluxWriteDataCommand())
	rootCmd.AddCommand(AbapEnvironmentRunAUnitTestCommand())
	rootCmd.AddCommand(CheckStepActiveCommand())
	rootCmd.AddCommand(GolangBuildCommand())
	rootCmd.AddCommand(ShellExecuteCommand())
	rootCmd.AddCommand(ApiProxyDownloadCommand())
	rootCmd.AddCommand(ApiKeyValueMapDownloadCommand())
	rootCmd.AddCommand(ApiProviderDownloadCommand())
	rootCmd.AddCommand(ApiProxyUploadCommand())
	rootCmd.AddCommand(GradleExecuteBuildCommand())
	rootCmd.AddCommand(ApiKeyValueMapUploadCommand())
	rootCmd.AddCommand(PythonBuildCommand())
	rootCmd.AddCommand(AwsS3UploadCommand())

	addRootFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		log.Entry().WithError(err).Fatal("configuration error")
	}
}

func addRootFlags(rootCmd *cobra.Command) {
	var provider orchestrator.OrchestratorSpecificConfigProviding
	var err error

	provider, err = orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().Error(err)
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}

	rootCmd.PersistentFlags().StringVar(&GeneralConfig.CorrelationID, "correlationID", provider.GetBuildURL(), "ID for unique identification of a pipeline run")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.CustomConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&GeneralConfig.GitHubTokens, "gitHubTokens", AccessTokensFromEnvJSON(os.Getenv("PIPER_gitHubTokens")), "List of entries in form of <hostname>:<token> to allow GitHub token authentication for downloading config / defaults")
	rootCmd.PersistentFlags().StringSliceVar(&GeneralConfig.DefaultConfig, "defaultConfig", []string{".pipeline/defaults.yaml"}, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().BoolVar(&GeneralConfig.IgnoreCustomDefaults, "ignoreCustomDefaults", false, "Disables evaluation of the parameter 'customDefaults' in the pipeline configuration file")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.ParametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.HookConfig.ANSConfig.ServiceKey, "ansServiceKey", os.Getenv("PIPER_ansServiceKey"), "Service Key JSON needed for ANS")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.HookConfig.ANSConfig.EventTemplate, "ansEventTemplate", os.Getenv("PIPER_ansEventTemplate"), "Optional ANS event template JSON string")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.EnvRootPath, "envRootPath", ".pipeline", "Root path to Piper pipeline shared environments")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StageName, "stageName", "", "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.StepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVar(&GeneralConfig.NoTelemetry, "noTelemetry", false, "Disables telemetry reporting")
	rootCmd.PersistentFlags().BoolVarP(&GeneralConfig.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.LogFormat, "logFormat", "default", "Log format to use. Options: default, timestamp, plain, full.")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.VaultServerURL, "vaultServerUrl", "", "The Vault server which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.VaultNamespace, "vaultNamespace", "", "The Vault namespace which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.VaultPath, "vaultPath", "", "The path which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.GCPJsonKeyFilePath, "gcpJsonKeyFilePath", "", "File path to Google Cloud Platform JSON key file")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.GCSFolderPath, "gcsFolderPath", "", "GCS folder path. One of the components of GCS target folder")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.GCSBucketId, "gcsBucketId", "", "Bucket name for Google Cloud Storage")
	rootCmd.PersistentFlags().StringVar(&GeneralConfig.GCSSubFolder, "gcsSubFolder", "", "Used to logically separate results of the same step result type")

}

// ResolveAccessTokens reads a list of tokens in format host:token passed via command line
// and transfers this into a map as a more consumable format.
func ResolveAccessTokens(tokenList []string) map[string]string {
	tokenMap := map[string]string{}
	for _, tokenEntry := range tokenList {
		log.Entry().Debugf("processing token %v", tokenEntry)
		parts := strings.Split(tokenEntry, ":")
		if len(parts) != 2 {
			log.Entry().Warningf("wrong format for access token %v", tokenEntry)
		} else {
			tokenMap[parts[0]] = parts[1]
		}
	}
	return tokenMap
}

// AccessTokensFromEnvJSON resolves access tokens when passed as JSON in an environment variable
func AccessTokensFromEnvJSON(env string) []string {
	accessTokens := []string{}
	if len(env) == 0 {
		return accessTokens
	}
	err := json.Unmarshal([]byte(env), &accessTokens)
	if err != nil {
		log.Entry().Infof("Token json '%v' has wrong format.", env)
	}
	return accessTokens
}

// initStageName initializes GeneralConfig.StageName from either GeneralConfig.ParametersJSON
// or the environment variable (orchestrator specific), unless it has been provided as command line option.
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
	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().WithError(err).Warning("Cannot infer stage name from CI environment")
	} else {
		stageNameSource = "env variable"
		GeneralConfig.StageName = provider.GetStageName()
	}

	if len(GeneralConfig.ParametersJSON) == 0 {
		return
	}

	var params map[string]interface{}
	err = json.Unmarshal([]byte(GeneralConfig.ParametersJSON), &params)
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
func PrepareConfig(cmd *cobra.Command, metadata *config.StepData, stepName string, options interface{}, openFile func(s string, t map[string]string) (io.ReadCloser, error)) error {

	log.SetFormatter(GeneralConfig.LogFormat)

	initStageName(true)

	filters := metadata.GetParameterFilters()

	// add telemetry parameter "collectTelemetryData" to ALL, GENERAL and PARAMETER filters
	filters.All = append(filters.All, "collectTelemetryData")
	filters.General = append(filters.General, "collectTelemetryData")
	filters.Parameters = append(filters.Parameters, "collectTelemetryData")

	envParams := metadata.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	reportingEnvParams := config.ReportingParameters.GetResourceParameters(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	resourceParams := mergeResourceParameters(envParams, reportingEnvParams)

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
	if GeneralConfig.VaultToken == "" {
		GeneralConfig.VaultToken = os.Getenv("PIPER_vaultToken")
	}
	myConfig.SetVaultCredentials(GeneralConfig.VaultRoleID, GeneralConfig.VaultRoleSecretID, GeneralConfig.VaultToken)

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
				if customConfig, err = openFile(projectConfigFile, GeneralConfig.GitHubAccessTokens); err != nil {
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
			fc, err := openFile(projectDefaultFile, GeneralConfig.GitHubAccessTokens)
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
		stepConfig, err = myConfig.GetStepConfig(flagValues, GeneralConfig.ParametersJSON, customConfig, defaultConfig, GeneralConfig.IgnoreCustomDefaults, filters, *metadata, resourceParams, GeneralConfig.StageName, stepName)
		if verbose, ok := stepConfig.Config["verbose"].(bool); ok && verbose {
			log.SetVerbose(verbose)
			GeneralConfig.Verbose = verbose
		} else if !ok && stepConfig.Config["verbose"] != nil {
			log.Entry().Warnf("invalid value for parameter verbose: '%v'", stepConfig.Config["verbose"])
		}
		if err != nil {
			return errors.Wrap(err, "retrieving step configuration failed")
		}
	}

	if fmt.Sprintf("%v", stepConfig.Config["collectTelemetryData"]) == "false" {
		GeneralConfig.NoTelemetry = true
	}

	stepConfig.Config = checkTypes(stepConfig.Config, options)
	confJSON, _ := json.Marshal(stepConfig.Config)
	_ = json.Unmarshal(confJSON, &options)

	config.MarkFlagsWithValue(cmd, stepConfig)

	retrieveHookConfig(stepConfig.HookConfig, &GeneralConfig.HookConfig)

	if GeneralConfig.GCPJsonKeyFilePath == "" {
		GeneralConfig.GCPJsonKeyFilePath, _ = stepConfig.Config["gcpJsonKeyFilePath"].(string)
	}
	if GeneralConfig.GCSFolderPath == "" {
		GeneralConfig.GCSFolderPath, _ = stepConfig.Config["gcsFolderPath"].(string)
	}
	if GeneralConfig.GCSBucketId == "" {
		GeneralConfig.GCSBucketId, _ = stepConfig.Config["gcsBucketId"].(string)
	}
	if GeneralConfig.GCSSubFolder == "" {
		GeneralConfig.GCSSubFolder, _ = stepConfig.Config["gcsSubFolder"].(string)
	}
	return nil
}

func retrieveHookConfig(source map[string]interface{}, target *HookConfiguration) {
	if source != nil {
		log.Entry().Info("Retrieving hook configuration")
		b, err := json.Marshal(source)
		if err != nil {
			log.Entry().Warningf("Failed to marshal source hook configuration: %v", err)
		}
		err = json.Unmarshal(b, target)
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

func mergeResourceParameters(resParams ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range resParams {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
