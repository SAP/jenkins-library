// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type sonarExecuteScanOptions struct {
	Instance                  string   `json:"instance,omitempty"`
	Proxy                     string   `json:"proxy,omitempty"`
	ServerURL                 string   `json:"serverUrl,omitempty"`
	Token                     string   `json:"token,omitempty"`
	Organization              string   `json:"organization,omitempty"`
	CustomTLSCertificateLinks []string `json:"customTlsCertificateLinks,omitempty"`
	SonarScannerDownloadURL   string   `json:"sonarScannerDownloadUrl,omitempty"`
	VersioningModel           string   `json:"versioningModel,omitempty" validate:"possible-values=major major-minor semantic full"`
	Version                   string   `json:"version,omitempty"`
	CustomScanVersion         string   `json:"customScanVersion,omitempty"`
	ProjectKey                string   `json:"projectKey,omitempty"`
	CoverageExclusions        []string `json:"coverageExclusions,omitempty"`
	InferJavaBinaries         bool     `json:"inferJavaBinaries,omitempty"`
	InferJavaLibraries        bool     `json:"inferJavaLibraries,omitempty"`
	Options                   []string `json:"options,omitempty"`
	WaitForQualityGate        bool     `json:"waitForQualityGate,omitempty"`
	BranchName                string   `json:"branchName,omitempty"`
	InferBranchName           bool     `json:"inferBranchName,omitempty"`
	ChangeID                  string   `json:"changeId,omitempty"`
	ChangeBranch              string   `json:"changeBranch,omitempty"`
	ChangeTarget              string   `json:"changeTarget,omitempty"`
	PullRequestProvider       string   `json:"pullRequestProvider,omitempty" validate:"possible-values=GitHub"`
	Owner                     string   `json:"owner,omitempty"`
	Repository                string   `json:"repository,omitempty"`
	GithubToken               string   `json:"githubToken,omitempty"`
	DisableInlineComments     bool     `json:"disableInlineComments,omitempty"`
	LegacyPRHandling          bool     `json:"legacyPRHandling,omitempty"`
	GithubAPIURL              string   `json:"githubApiUrl,omitempty"`
	M2Path                    string   `json:"m2Path,omitempty"`
}

type sonarExecuteScanReports struct {
}

func (p *sonarExecuteScanReports) persist(stepConfig sonarExecuteScanOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/sonarscan.json", ParamRef: "", StepResultType: "sonarqube"},
		{FilePattern: "**/sonarscan-result.json", ParamRef: "", StepResultType: "sonarqube"},
	}
	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcpJsonKeyFilePath, Modified: false},
	}
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
		return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
	}
}

type sonarExecuteScanInflux struct {
	step_data struct {
		fields struct {
			sonar bool
		}
		tags struct {
		}
	}
	sonarqube_data struct {
		fields struct {
			blocker_issues  int
			critical_issues int
			major_issues    int
			minor_issues    int
			info_issues     int
		}
		tags struct {
		}
	}
}

func (i *sonarExecuteScanInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "sonar", value: i.step_data.fields.sonar},
		{valType: config.InfluxField, measurement: "sonarqube_data", name: "blocker_issues", value: i.sonarqube_data.fields.blocker_issues},
		{valType: config.InfluxField, measurement: "sonarqube_data", name: "critical_issues", value: i.sonarqube_data.fields.critical_issues},
		{valType: config.InfluxField, measurement: "sonarqube_data", name: "major_issues", value: i.sonarqube_data.fields.major_issues},
		{valType: config.InfluxField, measurement: "sonarqube_data", name: "minor_issues", value: i.sonarqube_data.fields.minor_issues},
		{valType: config.InfluxField, measurement: "sonarqube_data", name: "info_issues", value: i.sonarqube_data.fields.info_issues},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Influx environment")
	}
}

// SonarExecuteScanCommand Executes the Sonar scanner
func SonarExecuteScanCommand() *cobra.Command {
	const STEP_NAME = "sonarExecuteScan"

	metadata := sonarExecuteScanMetadata()
	var stepConfig sonarExecuteScanOptions
	var startTime time.Time
	var reports sonarExecuteScanReports
	var influx sonarExecuteScanInflux
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createSonarExecuteScanCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes the Sonar scanner",
		Long:  `The step executes the [sonar-scanner](https://docs.sonarqube.org/display/SCAN/Analyzing+with+SonarQube+Scanner) cli command to scan the defined sources and publish the results to a SonarQube instance.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.Token)
			log.RegisterSecret(stepConfig.GithubToken)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 || len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			sonarExecuteScan(stepConfig, &stepTelemetryData, &influx)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addSonarExecuteScanFlags(createSonarExecuteScanCmd, &stepConfig)
	return createSonarExecuteScanCmd
}

func addSonarExecuteScanFlags(cmd *cobra.Command, stepConfig *sonarExecuteScanOptions) {
	cmd.Flags().StringVar(&stepConfig.Instance, "instance", os.Getenv("PIPER_instance"), "Jenkins only: The name of the SonarQube instance defined in the Jenkins settings. DEPRECATED: use serverUrl parameter instead")
	cmd.Flags().StringVar(&stepConfig.Proxy, "proxy", os.Getenv("PIPER_proxy"), "Proxy URL to be used for communication with the SonarQube instance.")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", os.Getenv("PIPER_serverUrl"), "The URL to the Sonar backend. The serverUrl parameter requires the [`instance`](#instance) parameter to be explicitly set to an empty string, as it will have no effect otherwise.")
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "Token used to authenticate with the Sonar Server.")
	cmd.Flags().StringVar(&stepConfig.Organization, "organization", os.Getenv("PIPER_organization"), "SonarCloud.io only: Organization that the project will be assigned to in SonarCloud.io.")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List of download links to custom TLS certificates. This is required to ensure trusted connections to instances with custom certificates.")
	cmd.Flags().StringVar(&stepConfig.SonarScannerDownloadURL, "sonarScannerDownloadUrl", `https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.8.0.2856-linux.zip`, "URL to the sonar-scanner-cli archive.")
	cmd.Flags().StringVar(&stepConfig.VersioningModel, "versioningModel", `major`, "The versioning model used for the version when reporting the results for the project.")
	cmd.Flags().StringVar(&stepConfig.Version, "version", os.Getenv("PIPER_version"), "The project version that is reported to SonarQube.")
	cmd.Flags().StringVar(&stepConfig.CustomScanVersion, "customScanVersion", os.Getenv("PIPER_customScanVersion"), "A custom version used along with the uploaded scan results.")
	cmd.Flags().StringVar(&stepConfig.ProjectKey, "projectKey", os.Getenv("PIPER_projectKey"), "The project key identifies the project in SonarQube.")
	cmd.Flags().StringSliceVar(&stepConfig.CoverageExclusions, "coverageExclusions", []string{}, "A list of patterns that should be excluded from the coverage scan.")
	cmd.Flags().BoolVar(&stepConfig.InferJavaBinaries, "inferJavaBinaries", false, "Find the location of generated Java class files in all modules and pass the option `sonar.java.binaries to the sonar tool.")
	cmd.Flags().BoolVar(&stepConfig.InferJavaLibraries, "inferJavaLibraries", false, "If the parameter `m2Path` is configured for the step `mavenExecute` in the general section of the configuration, pass it as option `sonar.java.libraries` to the sonar tool.")
	cmd.Flags().StringSliceVar(&stepConfig.Options, "options", []string{}, "A list of options which are passed to the sonar-scanner.")
	cmd.Flags().BoolVar(&stepConfig.WaitForQualityGate, "waitForQualityGate", false, "Whether the scan should wait for and consider the result of the quality gate.")
	cmd.Flags().StringVar(&stepConfig.BranchName, "branchName", os.Getenv("PIPER_branchName"), "Non-Pull-Request only: Name of the SonarQube branch that should be used to report findings to. Automatically inferred from environment variables on supported orchestrators if `inferBranchName` is set to true.")
	cmd.Flags().BoolVar(&stepConfig.InferBranchName, "inferBranchName", false, "Whether to infer the `branchName` parameter automatically based on the orchestrator-specific environment variable in runs of the pipeline.")
	cmd.Flags().StringVar(&stepConfig.ChangeID, "changeId", os.Getenv("PIPER_changeId"), "Pull-Request only: The id of the pull-request. Automatically inferred from environment variables on supported orchestrators.")
	cmd.Flags().StringVar(&stepConfig.ChangeBranch, "changeBranch", os.Getenv("PIPER_changeBranch"), "Pull-Request only: The name of the pull-request branch. Automatically inferred from environment variables on supported orchestrators.")
	cmd.Flags().StringVar(&stepConfig.ChangeTarget, "changeTarget", os.Getenv("PIPER_changeTarget"), "Pull-Request only: The name of the base branch. Automatically inferred from environment variables on supported orchestrators.")
	cmd.Flags().StringVar(&stepConfig.PullRequestProvider, "pullRequestProvider", `GitHub`, "Pull-Request only: The scm provider.")
	cmd.Flags().StringVar(&stepConfig.Owner, "owner", os.Getenv("PIPER_owner"), "Pull-Request only: The owner of the scm repository.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Pull-Request only: The scm repository.")
	cmd.Flags().StringVar(&stepConfig.GithubToken, "githubToken", os.Getenv("PIPER_githubToken"), "Pull-Request only: Token for Github to set status on the Pull-Request.")
	cmd.Flags().BoolVar(&stepConfig.DisableInlineComments, "disableInlineComments", false, "Pull-Request only: Disables the pull-request decoration with inline comments. DEPRECATED: only supported in SonarQube < 7.2")
	cmd.Flags().BoolVar(&stepConfig.LegacyPRHandling, "legacyPRHandling", false, "Pull-Request only: Activates the pull-request handling using the [GitHub Plugin](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin). DEPRECATED: only supported in SonarQube < 7.2")
	cmd.Flags().StringVar(&stepConfig.GithubAPIURL, "githubApiUrl", `https://api.github.com`, "Pull-Request only: The URL to the Github API. See [GitHub plugin docs](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin#GitHubPlugin-Usage) DEPRECATED: only supported in SonarQube < 7.2")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")

}

// retrieve step metadata
func sonarExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "sonarExecuteScan",
			Aliases:     []config.Alias{},
			Description: "Executes the Sonar scanner",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "sonarTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing the token used to authenticate with the Sonar Server.", Type: "jenkins"},
					{Name: "githubTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing the token used to authenticate with the Github Server.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "instance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_instance"),
					},
					{
						Name:        "proxy",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS", "STAGES"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_proxy"),
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "host"}, {Name: "sonarServerUrl"}},
						Default:     os.Getenv("PIPER_serverUrl"),
					},
					{
						Name: "token",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "sonarVaultSecretName",
								Type:    "vaultSecret",
								Default: "sonar",
							},

							{
								Name: "sonarTokenCredentialsId",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "sonarToken"}},
						Default:   os.Getenv("PIPER_token"),
					},
					{
						Name:        "organization",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_organization"),
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "sonarScannerDownloadUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.8.0.2856-linux.zip`,
					},
					{
						Name:        "versioningModel",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STAGES", "STEPS", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `major`,
					},
					{
						Name: "version",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "projectVersion", Deprecated: true}},
						Default:   os.Getenv("PIPER_version"),
					},
					{
						Name:        "customScanVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STAGES", "STEPS", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_customScanVersion"),
					},
					{
						Name:        "projectKey",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_projectKey"),
					},
					{
						Name:        "coverageExclusions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "inferJavaBinaries",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "inferJavaLibraries",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "options",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "sonarProperties", Deprecated: true}},
						Default:     []string{},
					},
					{
						Name:        "waitForQualityGate",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "branchName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_branchName"),
					},
					{
						Name:        "inferBranchName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "changeId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_changeId"),
					},
					{
						Name:        "changeBranch",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_changeBranch"),
					},
					{
						Name:        "changeTarget",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_changeTarget"),
					},
					{
						Name:        "pullRequestProvider",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `GitHub`,
					},
					{
						Name: "owner",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/owner",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "githubOrg"}},
						Default:   os.Getenv("PIPER_owner"),
					},
					{
						Name: "repository",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/repository",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "githubRepo"}},
						Default:   os.Getenv("PIPER_repository"),
					},
					{
						Name: "githubToken",
						ResourceRef: []config.ResourceReference{
							{
								Name: "githubTokenCredentialsId",
								Type: "secret",
							},

							{
								Name:    "githubVaultSecretName",
								Type:    "vaultSecret",
								Default: "github",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "access_token"}},
						Default:   os.Getenv("PIPER_githubToken"),
					},
					{
						Name:        "disableInlineComments",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "legacyPRHandling",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "githubApiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `https://api.github.com`,
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
						Default:     os.Getenv("PIPER_m2Path"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "sonar", Image: "sonarsource/sonar-scanner-cli:5.0", Options: []config.Option{{Name: "-u", Value: "0"}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/sonarscan.json", "type": "sonarqube"},
							{"filePattern": "**/sonarscan-result.json", "type": "sonarqube"},
						},
					},
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"name": "step_data", "fields": []map[string]string{{"name": "sonar"}}},
							{"name": "sonarqube_data", "fields": []map[string]string{{"name": "blocker_issues"}, {"name": "critical_issues"}, {"name": "major_issues"}, {"name": "minor_issues"}, {"name": "info_issues"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
