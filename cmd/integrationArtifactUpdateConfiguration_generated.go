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
	"github.com/spf13/cobra"
)

type integrationArtifactUpdateConfigurationOptions struct {
	APIServiceKey          string `json:"apiServiceKey,omitempty"`
	IntegrationFlowID      string `json:"integrationFlowId,omitempty"`
	IntegrationFlowVersion string `json:"integrationFlowVersion,omitempty"`
	Platform               string `json:"platform,omitempty"`
	ParameterKey           string `json:"parameterKey,omitempty"`
	ParameterValue         string `json:"parameterValue,omitempty"`
}

// IntegrationArtifactUpdateConfigurationCommand Update integration flow Configuration parameter
func IntegrationArtifactUpdateConfigurationCommand() *cobra.Command {
	const STEP_NAME = "integrationArtifactUpdateConfiguration"

	metadata := integrationArtifactUpdateConfigurationMetadata()
	var stepConfig integrationArtifactUpdateConfigurationOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createIntegrationArtifactUpdateConfigurationCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Update integration flow Configuration parameter",
		Long:  `With this step you can update the value for a configuration parameters of a designtime integration flow using the OData API. Learn more about the SAP Cloud Integration remote API for configuration update of the integration flow parameter [here](https://help.sap.com/viewer/368c481cd6954bdfa5d0435479fd4eaf/Cloud/en-US/d1679a80543f46509a7329243b595bdb.html).`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.APIServiceKey)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			integrationArtifactUpdateConfiguration(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addIntegrationArtifactUpdateConfigurationFlags(createIntegrationArtifactUpdateConfigurationCmd, &stepConfig)
	return createIntegrationArtifactUpdateConfigurationCmd
}

func addIntegrationArtifactUpdateConfigurationFlags(cmd *cobra.Command, stepConfig *integrationArtifactUpdateConfigurationOptions) {
	cmd.Flags().StringVar(&stepConfig.APIServiceKey, "apiServiceKey", os.Getenv("PIPER_apiServiceKey"), "Service key JSON string to access the Cloud Integration API")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowID, "integrationFlowId", os.Getenv("PIPER_integrationFlowId"), "Specifies the ID of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowVersion, "integrationFlowVersion", os.Getenv("PIPER_integrationFlowVersion"), "Specifies the version of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.Platform, "platform", os.Getenv("PIPER_platform"), "Specifies the running platform of the SAP Cloud platform integraion service")
	cmd.Flags().StringVar(&stepConfig.ParameterKey, "parameterKey", os.Getenv("PIPER_parameterKey"), "Specifies the externalized parameter name.")
	cmd.Flags().StringVar(&stepConfig.ParameterValue, "parameterValue", os.Getenv("PIPER_parameterValue"), "Specifies the externalized parameter value.")

	cmd.MarkFlagRequired("apiServiceKey")
	cmd.MarkFlagRequired("integrationFlowId")
	cmd.MarkFlagRequired("integrationFlowVersion")
	cmd.MarkFlagRequired("parameterKey")
	cmd.MarkFlagRequired("parameterValue")
}

// retrieve step metadata
func integrationArtifactUpdateConfigurationMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "integrationArtifactUpdateConfiguration",
			Aliases:     []config.Alias{},
			Description: "Update integration flow Configuration parameter",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "cpiApiServiceKeyCredentialsId", Description: "Jenkins credential ID for secret text containing the service key to the SAP Cloud Integration API", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "apiServiceKey",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cpiApiServiceKeyCredentialsId",
								Param: "apiServiceKey",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_apiServiceKey"),
					},
					{
						Name:        "integrationFlowId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_integrationFlowId"),
					},
					{
						Name:        "integrationFlowVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_integrationFlowVersion"),
					},
					{
						Name:        "platform",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_platform"),
					},
					{
						Name:        "parameterKey",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_parameterKey"),
					},
					{
						Name:        "parameterValue",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_parameterValue"),
					},
				},
			},
		},
	}
	return theMetaData
}
