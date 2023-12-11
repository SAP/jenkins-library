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

type codeqlExecuteScanOptions struct {
	GithubToken                 string `json:"githubToken,omitempty"`
	BuildTool                   string `json:"buildTool,omitempty" validate:"possible-values=custom maven golang npm pip yarn"`
	BuildCommand                string `json:"buildCommand,omitempty"`
	Language                    string `json:"language,omitempty"`
	ModulePath                  string `json:"modulePath,omitempty"`
	Database                    string `json:"database,omitempty"`
	QuerySuite                  string `json:"querySuite,omitempty"`
	UploadResults               bool   `json:"uploadResults,omitempty"`
	SarifCheckMaxRetries        int    `json:"sarifCheckMaxRetries,omitempty"`
	SarifCheckRetryInterval     int    `json:"sarifCheckRetryInterval,omitempty"`
	TargetGithubRepoURL         string `json:"targetGithubRepoURL,omitempty"`
	TargetGithubBranchName      string `json:"targetGithubBranchName,omitempty"`
	Threads                     string `json:"threads,omitempty"`
	Ram                         string `json:"ram,omitempty"`
	AnalyzedRef                 string `json:"analyzedRef,omitempty"`
	Repository                  string `json:"repository,omitempty"`
	CommitID                    string `json:"commitId,omitempty"`
	VulnerabilityThresholdTotal int    `json:"vulnerabilityThresholdTotal,omitempty"`
	CheckForCompliance          bool   `json:"checkForCompliance,omitempty"`
	ProjectSettingsFile         string `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile          string `json:"globalSettingsFile,omitempty"`
}

type codeqlExecuteScanInflux struct {
	step_data struct {
		fields struct {
			codeql bool
		}
		tags struct {
		}
	}
	codeql_data struct {
		fields struct {
			projectID         int64
			projectName       string
			projectVersion    string
			projectVersionID  int64
			violations        int
			corporateTotal    int
			corporateAudited  int
			auditAllTotal     int
			auditAllAudited   int
			spotChecksTotal   int
			spotChecksAudited int
			spotChecksGap     int
			suspicious        int
			exploitable       int
			suppressed        int
		}
		tags struct {
		}
	}
}

func (i *codeqlExecuteScanInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "codeql", value: i.step_data.fields.codeql},
		{valType: config.InfluxField, measurement: "codeql_data", name: "projectID", value: i.codeql_data.fields.projectID},
		{valType: config.InfluxField, measurement: "codeql_data", name: "projectName", value: i.codeql_data.fields.projectName},
		{valType: config.InfluxField, measurement: "codeql_data", name: "projectVersion", value: i.codeql_data.fields.projectVersion},
		{valType: config.InfluxField, measurement: "codeql_data", name: "projectVersionId", value: i.codeql_data.fields.projectVersionID},
		{valType: config.InfluxField, measurement: "codeql_data", name: "violations", value: i.codeql_data.fields.violations},
		{valType: config.InfluxField, measurement: "codeql_data", name: "corporateTotal", value: i.codeql_data.fields.corporateTotal},
		{valType: config.InfluxField, measurement: "codeql_data", name: "corporateAudited", value: i.codeql_data.fields.corporateAudited},
		{valType: config.InfluxField, measurement: "codeql_data", name: "auditAllTotal", value: i.codeql_data.fields.auditAllTotal},
		{valType: config.InfluxField, measurement: "codeql_data", name: "auditAllAudited", value: i.codeql_data.fields.auditAllAudited},
		{valType: config.InfluxField, measurement: "codeql_data", name: "spotChecksTotal", value: i.codeql_data.fields.spotChecksTotal},
		{valType: config.InfluxField, measurement: "codeql_data", name: "spotChecksAudited", value: i.codeql_data.fields.spotChecksAudited},
		{valType: config.InfluxField, measurement: "codeql_data", name: "spotChecksGap", value: i.codeql_data.fields.spotChecksGap},
		{valType: config.InfluxField, measurement: "codeql_data", name: "suspicious", value: i.codeql_data.fields.suspicious},
		{valType: config.InfluxField, measurement: "codeql_data", name: "exploitable", value: i.codeql_data.fields.exploitable},
		{valType: config.InfluxField, measurement: "codeql_data", name: "suppressed", value: i.codeql_data.fields.suppressed},
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

type codeqlExecuteScanReports struct {
}

func (p *codeqlExecuteScanReports) persist(stepConfig codeqlExecuteScanOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/*.csv", ParamRef: "", StepResultType: "codeql"},
		{FilePattern: "**/*.sarif", ParamRef: "", StepResultType: "codeql"},
		{FilePattern: "**/toolrun_codeql_*.json", ParamRef: "", StepResultType: "codeql"},
		{FilePattern: "**/piper_codeql_report.json", ParamRef: "", StepResultType: "codeql"},
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

// CodeqlExecuteScanCommand This step executes a codeql scan on the specified project to perform static code analysis and check the source code for security flaws.
func CodeqlExecuteScanCommand() *cobra.Command {
	const STEP_NAME = "codeqlExecuteScan"

	metadata := codeqlExecuteScanMetadata()
	var stepConfig codeqlExecuteScanOptions
	var startTime time.Time
	var influx codeqlExecuteScanInflux
	var reports codeqlExecuteScanReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createCodeqlExecuteScanCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step executes a codeql scan on the specified project to perform static code analysis and check the source code for security flaws.",
		Long: `This step executes a codeql scan on the specified project to perform static code analysis and check the source code for security flaws.

The codeql step triggers a scan locally on your orchestrator (e.g. Jenkins) within a docker container so finally you have to supply a docker image with codeql
and Java plus Maven.`,
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
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
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
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			codeqlExecuteScan(stepConfig, &stepTelemetryData, &influx)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addCodeqlExecuteScanFlags(createCodeqlExecuteScanCmd, &stepConfig)
	return createCodeqlExecuteScanCmd
}

func addCodeqlExecuteScanFlags(cmd *cobra.Command, stepConfig *codeqlExecuteScanOptions) {
	cmd.Flags().StringVar(&stepConfig.GithubToken, "githubToken", os.Getenv("PIPER_githubToken"), "GitHub personal access token in plain text. NEVER set this parameter in a file commited to a source code repository. This parameter is intended to be used from the command line or set securely via the environment variable listed below. In most pipeline use-cases, you should instead either store the token in Vault (where it can be automatically retrieved by the step from one of the paths listed below) or store it as a Jenkins secret and configure the secret's id via the `githubTokenCredentialsId` parameter.")
	cmd.Flags().StringVar(&stepConfig.BuildTool, "buildTool", `maven`, "Defines the build tool which is used for building the project.")
	cmd.Flags().StringVar(&stepConfig.BuildCommand, "buildCommand", os.Getenv("PIPER_buildCommand"), "Command to build the project")
	cmd.Flags().StringVar(&stepConfig.Language, "language", os.Getenv("PIPER_language"), "The programming language used to analyze.")
	cmd.Flags().StringVar(&stepConfig.ModulePath, "modulePath", `./`, "Allows providing the path for the module to scan")
	cmd.Flags().StringVar(&stepConfig.Database, "database", `codeqlDB`, "Path to the CodeQL database to create. This directory will be created, and must not already exist.")
	cmd.Flags().StringVar(&stepConfig.QuerySuite, "querySuite", os.Getenv("PIPER_querySuite"), "The name of a CodeQL query suite. If omitted, the default query suite for the language of the database being analyzed will be used.")
	cmd.Flags().BoolVar(&stepConfig.UploadResults, "uploadResults", false, "Allows you to upload codeql SARIF results to your github project. You will need to set githubToken for this.")
	cmd.Flags().IntVar(&stepConfig.SarifCheckMaxRetries, "sarifCheckMaxRetries", 10, "Maximum number of retries when waiting for the server to finish processing the SARIF upload.")
	cmd.Flags().IntVar(&stepConfig.SarifCheckRetryInterval, "sarifCheckRetryInterval", 30, "Interval in seconds between retries when waiting for the server to finish processing the SARIF upload.")
	cmd.Flags().StringVar(&stepConfig.TargetGithubRepoURL, "targetGithubRepoURL", os.Getenv("PIPER_targetGithubRepoURL"), "")
	cmd.Flags().StringVar(&stepConfig.TargetGithubBranchName, "targetGithubBranchName", os.Getenv("PIPER_targetGithubBranchName"), "")
	cmd.Flags().StringVar(&stepConfig.Threads, "threads", `0`, "Use this many threads for the codeql operations.")
	cmd.Flags().StringVar(&stepConfig.Ram, "ram", os.Getenv("PIPER_ram"), "Use this much ram (MB) for the codeql operations.")
	cmd.Flags().StringVar(&stepConfig.AnalyzedRef, "analyzedRef", os.Getenv("PIPER_analyzedRef"), "Name of the ref that was analyzed.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "URL of the GitHub instance")
	cmd.Flags().StringVar(&stepConfig.CommitID, "commitId", os.Getenv("PIPER_commitId"), "SHA of commit that was analyzed.")
	cmd.Flags().IntVar(&stepConfig.VulnerabilityThresholdTotal, "vulnerabilityThresholdTotal", 0, "Threashold for maximum number of allowed vulnerabilities.")
	cmd.Flags().BoolVar(&stepConfig.CheckForCompliance, "checkForCompliance", false, "If set to true, the piper step checks for compliance based on vulnerability threadholds. Example - If total vulnerabilites are 10 and vulnerabilityThresholdTotal is set as 0, then the steps throws an compliance error.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path to the mvn settings file that should be used as global settings file.")

	cmd.MarkFlagRequired("buildTool")
}

// retrieve step metadata
func codeqlExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "codeqlExecuteScan",
			Aliases:     []config.Alias{},
			Description: "This step executes a codeql scan on the specified project to perform static code analysis and check the source code for security flaws.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "githubTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing token to authenticate to GitHub.", Type: "jenkins"},
				},
				Resources: []config.StepResources{
					{Name: "commonPipelineEnvironment"},
					{Name: "buildDescriptor", Type: "stash"},
					{Name: "tests", Type: "stash"},
				},
				Parameters: []config.StepParameters{
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
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "access_token"}},
						Default:   os.Getenv("PIPER_githubToken"),
					},
					{
						Name:        "buildTool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `maven`,
					},
					{
						Name:        "buildCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_buildCommand"),
					},
					{
						Name:        "language",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_language"),
					},
					{
						Name:        "modulePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `./`,
					},
					{
						Name:        "database",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `codeqlDB`,
					},
					{
						Name:        "querySuite",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_querySuite"),
					},
					{
						Name:        "uploadResults",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "sarifCheckMaxRetries",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     10,
					},
					{
						Name:        "sarifCheckRetryInterval",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     30,
					},
					{
						Name:        "targetGithubRepoURL",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_targetGithubRepoURL"),
					},
					{
						Name:        "targetGithubBranchName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_targetGithubBranchName"),
					},
					{
						Name:        "threads",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `0`,
					},
					{
						Name:        "ram",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_ram"),
					},
					{
						Name: "analyzedRef",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "git/ref",
							},
						},
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_analyzedRef"),
					},
					{
						Name: "repository",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "git/httpsUrl",
							},
						},
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "githubRepo"}},
						Default:   os.Getenv("PIPER_repository"),
					},
					{
						Name: "commitId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "git/remoteCommitId",
							},
						},
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_commitId"),
					},
					{
						Name:        "vulnerabilityThresholdTotal",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     0,
					},
					{
						Name:        "checkForCompliance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/projectSettingsFile"}},
						Default:     os.Getenv("PIPER_projectSettingsFile"),
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/globalSettingsFile"}},
						Default:     os.Getenv("PIPER_globalSettingsFile"),
					},
				},
			},
			Containers: []config.Container{
				{},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"name": "step_data", "fields": []map[string]string{{"name": "codeql"}}},
							{"name": "codeql_data", "fields": []map[string]string{{"name": "projectID"}, {"name": "projectName"}, {"name": "projectVersion"}, {"name": "projectVersionId"}, {"name": "violations"}, {"name": "corporateTotal"}, {"name": "corporateAudited"}, {"name": "auditAllTotal"}, {"name": "auditAllAudited"}, {"name": "spotChecksTotal"}, {"name": "spotChecksAudited"}, {"name": "spotChecksGap"}, {"name": "suspicious"}, {"name": "exploitable"}, {"name": "suppressed"}}},
						},
					},
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/*.csv", "type": "codeql"},
							{"filePattern": "**/*.sarif", "type": "codeql"},
							{"filePattern": "**/toolrun_codeql_*.json", "type": "codeql"},
							{"filePattern": "**/piper_codeql_report.json", "type": "codeql"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
