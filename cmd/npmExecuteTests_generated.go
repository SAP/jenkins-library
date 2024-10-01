// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type npmExecuteTestsOptions struct {
	InstallCommand            string `json:"installCommand,omitempty"`
	RunScript                 string `json:"runScript,omitempty"`
	AppURLs                   string `json:"appUrls,omitempty"`
	OnlyRunInProductiveBranch bool   `json:"onlyRunInProductiveBranch,omitempty"`
	ProductiveBranch          string `json:"productiveBranch,omitempty"`
	BaseURL                   string `json:"baseUrl,omitempty"`
	Wdi5                      bool   `json:"wdi5,omitempty"`
	CredentialsID             string `json:"credentialsId,omitempty"`
}

type npmExecuteTestsReports struct {
}

func (p *npmExecuteTestsReports) persist(stepConfig npmExecuteTestsOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/e2e-results.xml", ParamRef: "", StepResultType: "e2e"},
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

// NpmExecuteTestsCommand Executes end-to-end tests using npm
func NpmExecuteTestsCommand() *cobra.Command {
	const STEP_NAME = "npmExecuteTests"

	metadata := npmExecuteTestsMetadata()
	var stepConfig npmExecuteTestsOptions
	var startTime time.Time
	var reports npmExecuteTestsReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createNpmExecuteTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes end-to-end tests using npm",
		Long: `This step executes end-to-end tests in a Docker environment using npm.

The step spins up a Docker container based on the specified ` + "`" + `dockerImage` + "`" + ` and executes the ` + "`" + `installScript` + "`" + ` and ` + "`" + `runScript` + "`" + ` from ` + "`" + `package.json` + "`" + `.

The application URLs and credentials can be specified in ` + "`" + `appUrls` + "`" + ` and ` + "`" + `credentialsId` + "`" + ` respectively. If ` + "`" + `wdi5` + "`" + ` is set to ` + "`" + `true` + "`" + `, the step uses ` + "`" + `wdi5_username` + "`" + ` and ` + "`" + `wdi5_password` + "`" + ` for authentication.

The tests can be restricted to run only on the productive branch by setting ` + "`" + `onlyRunInProductiveBranch` + "`" + ` to ` + "`" + `true` + "`" + `.`,
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
			npmExecuteTests(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addNpmExecuteTestsFlags(createNpmExecuteTestsCmd, &stepConfig)
	return createNpmExecuteTestsCmd
}

func addNpmExecuteTestsFlags(cmd *cobra.Command, stepConfig *npmExecuteTestsOptions) {
	cmd.Flags().StringVar(&stepConfig.InstallCommand, "installCommand", `npm install`, "Command to be executed for installation. Defaults to `npm install`.")
	cmd.Flags().StringVar(&stepConfig.RunScript, "runScript", `wdi5`, "Script to be executed from package.json for running tests. Defaults to `wdi5`.")
	cmd.Flags().StringVar(&stepConfig.AppURLs, "appUrls", `[]`, "A JSON string containing an array of objects, each representing an application URL with associated credentials.\nEach object must have the following properties:\n- `url`: The URL of the application.\n- `username`: The username for accessing the application.\n- `password`: The password for accessing the application.\nThis parameter is used to securely pass multiple application URLs and their credentials from Vault.\n")
	cmd.Flags().BoolVar(&stepConfig.OnlyRunInProductiveBranch, "onlyRunInProductiveBranch", false, "Boolean   to indicate whether the step should only be executed in the productive branch or not.")
	cmd.Flags().StringVar(&stepConfig.ProductiveBranch, "productiveBranch", `main`, "The branch used as productive branch.")
	cmd.Flags().StringVar(&stepConfig.BaseURL, "baseUrl", `0.0.0.0`, "Base URL of the application to be tested.")
	cmd.Flags().BoolVar(&stepConfig.Wdi5, "wdi5", true, "Distinguish if these are wdi5 tests.")
	cmd.Flags().StringVar(&stepConfig.CredentialsID, "credentialsId", os.Getenv("PIPER_credentialsId"), "Credentials to access the application to be tested.")

	cmd.MarkFlagRequired("runScript")
}

// retrieve step metadata
func npmExecuteTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "npmExecuteTests",
			Aliases:     []config.Alias{},
			Description: "Executes end-to-end tests using npm",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "installCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `npm install`,
					},
					{
						Name:        "runScript",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `wdi5`,
					},
					{
						Name: "appUrls",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "appMetadataVaultSecretName",
								Type:    "vaultSecret",
								Default: "appMetadata",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   `[]`,
					},
					{
						Name:        "onlyRunInProductiveBranch",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "productiveBranch",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `main`,
					},
					{
						Name:        "baseUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `0.0.0.0`,
					},
					{
						Name:        "wdi5",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "credentialsId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_credentialsId"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "node", Image: "node:lts-buster", EnvVars: []config.EnvVar{{Name: "BASE_URL", Value: "${{params.baseUrl}}"}, {Name: "CREDENTIALS_ID", Value: "${{params.credentialsId}}"}}, WorkingDir: "/app"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/e2e-results.xml", "type": "e2e"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
