// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type integrationArtifactTransportOptions struct {
	CasServiceKey        string `json:"casServiceKey,omitempty"`
	IntegrationPackageID string `json:"integrationPackageId,omitempty"`
	ResourceID           string `json:"resourceID,omitempty"`
	Name                 string `json:"name,omitempty"`
	Version              string `json:"version,omitempty"`
}

// IntegrationArtifactTransportCommand Integration Package transport using the SAP Content Agent Service
func IntegrationArtifactTransportCommand() *cobra.Command {
	const STEP_NAME = "integrationArtifactTransport"

	metadata := integrationArtifactTransportMetadata()
	var stepConfig integrationArtifactTransportOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createIntegrationArtifactTransportCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Integration Package transport using the SAP Content Agent Service",
		Long:  `With this step you can trigger an Integration Package transport from SAP Integration Suite using SAP Content Agent Service and SAP Cloud Transport Management Service. For more information about doing an Integration Package transport using SAP Content Agent Service see the documentation [here](https://help.sap.com/docs/CONTENT_AGENT_SERVICE/ae1a4f2d150d468d9ff56e13f9898e07/8e274fdd41da45a69ff919c0af8c6127.html).`,
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
			log.RegisterSecret(stepConfig.CasServiceKey)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
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
			integrationArtifactTransport(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addIntegrationArtifactTransportFlags(createIntegrationArtifactTransportCmd, &stepConfig)
	return createIntegrationArtifactTransportCmd
}

func addIntegrationArtifactTransportFlags(cmd *cobra.Command, stepConfig *integrationArtifactTransportOptions) {
	cmd.Flags().StringVar(&stepConfig.CasServiceKey, "casServiceKey", os.Getenv("PIPER_casServiceKey"), "Service key JSON string to access the CAS service instance")
	cmd.Flags().StringVar(&stepConfig.IntegrationPackageID, "integrationPackageId", os.Getenv("PIPER_integrationPackageId"), "Specifies the ID of the integration package artifact.")
	cmd.Flags().StringVar(&stepConfig.ResourceID, "resourceID", os.Getenv("PIPER_resourceID"), "Specifies the technical ID of the integration package artifact.")
	cmd.Flags().StringVar(&stepConfig.Name, "name", os.Getenv("PIPER_name"), "Specifies the name of the integration package artifact.")
	cmd.Flags().StringVar(&stepConfig.Version, "version", os.Getenv("PIPER_version"), "Specifies the version of the Integration Package artifact.")

	cmd.MarkFlagRequired("casServiceKey")
	cmd.MarkFlagRequired("integrationPackageId")
	cmd.MarkFlagRequired("resourceID")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("version")
}

// retrieve step metadata
func integrationArtifactTransportMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "integrationArtifactTransport",
			Aliases:     []config.Alias{},
			Description: "Integration Package transport using the SAP Content Agent Service",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "casApiServiceKeyCredentialsId", Description: "Jenkins secret text credential ID containing the service key to the CAS service instance", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "casServiceKey",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "casApiServiceKeyCredentialsId",
								Param: "casServiceKey",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_casServiceKey"),
					},
					{
						Name:        "integrationPackageId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_integrationPackageId"),
					},
					{
						Name:        "resourceID",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_resourceID"),
					},
					{
						Name:        "name",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_name"),
					},
					{
						Name:        "version",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_version"),
					},
				},
			},
		},
	}
	return theMetaData
}
