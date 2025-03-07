// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type onapsisExecuteScanOptions struct {
	ScanServiceURL         string `json:"scanServiceUrl,omitempty"`
	ScanGitURL             string `json:"scanGitUrl,omitempty"`
	OnapsisUsername        string `json:"onapsisUsername,omitempty"`
	OnapsisPassword        string `json:"onapsisPassword,omitempty"`
	AccessToken            string `json:"accessToken,omitempty"`
	AppType                string `json:"appType,omitempty" validate:"possible-values=ABAP SAPUI5"`
	FailOnMandatoryFinding bool   `json:"failOnMandatoryFinding,omitempty"`
	FailOnOptionalFinding  bool   `json:"failOnOptionalFinding,omitempty"`
	DebugMode              bool   `json:"debugMode,omitempty"`
}

// OnapsisExecuteScanCommand Execute a scan with Onapsis Control
func OnapsisExecuteScanCommand() *cobra.Command {
	const STEP_NAME = "onapsisExecuteScan"

	metadata := onapsisExecuteScanMetadata()
	var stepConfig onapsisExecuteScanOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createOnapsisExecuteScanCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Execute a scan with Onapsis Control",
		Long:  `This step executes a scan with Onapsis Control.`,
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
			log.RegisterSecret(stepConfig.AccessToken)

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
			vaultClient := config.GlobalVaultClient()
			if vaultClient != nil {
				defer vaultClient.MustRevokeToken()
			}

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
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
				if GeneralConfig.HookConfig.GCPPubSubConfig.Enabled {
					err := gcp.NewGcpPubsubClient(
						vaultClient,
						GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
						GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.OIDCConfig.RoleID,
					).Publish(GeneralConfig.HookConfig.GCPPubSubConfig.Topic, telemetryClient.GetDataBytes())
					if err != nil {
						log.Entry().WithError(err).Warn("event publish failed")
					}
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			onapsisExecuteScan(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addOnapsisExecuteScanFlags(createOnapsisExecuteScanCmd, &stepConfig)
	return createOnapsisExecuteScanCmd
}

func addOnapsisExecuteScanFlags(cmd *cobra.Command, stepConfig *onapsisExecuteScanOptions) {
	cmd.Flags().StringVar(&stepConfig.ScanServiceURL, "scanServiceUrl", os.Getenv("PIPER_scanServiceUrl"), "URL of the scan service")
	cmd.Flags().StringVar(&stepConfig.ScanGitURL, "scanGitUrl", os.Getenv("PIPER_scanGitUrl"), "The target git repo to scan")
	cmd.Flags().StringVar(&stepConfig.OnapsisUsername, "onapsisUsername", os.Getenv("PIPER_onapsisUsername"), "Onapsis username for JWT authentication")
	cmd.Flags().StringVar(&stepConfig.OnapsisPassword, "onapsisPassword", os.Getenv("PIPER_onapsisPassword"), "Onapsis password for JWT authentication")
	cmd.Flags().StringVar(&stepConfig.AccessToken, "accessToken", os.Getenv("PIPER_accessToken"), "Token used to authenticate with the Control Scan Service")
	cmd.Flags().StringVar(&stepConfig.AppType, "appType", `SAPUI5`, "Type of the application to be scanned")
	cmd.Flags().BoolVar(&stepConfig.FailOnMandatoryFinding, "failOnMandatoryFinding", true, "Fail the build if mandatory findings are detected")
	cmd.Flags().BoolVar(&stepConfig.FailOnOptionalFinding, "failOnOptionalFinding", false, "Fail the build if optional findings are detected")
	cmd.Flags().BoolVar(&stepConfig.DebugMode, "debugMode", false, "Enable debug mode for the scan")

	cmd.MarkFlagRequired("scanServiceUrl")
	cmd.MarkFlagRequired("onapsisUsername")
	cmd.MarkFlagRequired("onapsisPassword")
}

// retrieve step metadata
func onapsisExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "onapsisExecuteScan",
			Aliases:     []config.Alias{},
			Description: "Execute a scan with Onapsis Control",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "onapsisPassword", Description: "Onapsis password for JWT authentication", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "scanServiceUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name: "scanServiceUrl",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_scanServiceUrl"),
					},
					{
						Name:        "scanGitUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_scanGitUrl"),
					},
					{
						Name: "onapsisUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name: "onapsisUsername",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_onapsisUsername"),
					},
					{
						Name: "onapsisPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name: "onapsisPassword",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_onapsisPassword"),
					},
					{
						Name: "accessToken",
						ResourceRef: []config.ResourceReference{
							{
								Name: "onapsisTokenCredentialsId",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_accessToken"),
					},
					{
						Name:        "appType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `SAPUI5`,
					},
					{
						Name:        "failOnMandatoryFinding",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "failOnOptionalFinding",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "debugMode",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
				},
			},
		},
	}
	return theMetaData
}
