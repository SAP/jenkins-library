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

type abapLandscapePortalUpdateAddOnProductOptions struct {
	LandscapePortalAPIServiceKey string `json:"landscapePortalAPIServiceKey,omitempty"`
	AbapSystemNumber             string `json:"abapSystemNumber,omitempty"`
	AddonDescriptorFileName      string `json:"addonDescriptorFileName,omitempty"`
}

// AbapLandscapePortalUpdateAddOnProductCommand Update the AddOn product in SAP BTP ABAP Environment system of Landscape Portal
func AbapLandscapePortalUpdateAddOnProductCommand() *cobra.Command {
	const STEP_NAME = "abapLandscapePortalUpdateAddOnProduct"

	metadata := abapLandscapePortalUpdateAddOnProductMetadata()
	var stepConfig abapLandscapePortalUpdateAddOnProductOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createAbapLandscapePortalUpdateAddOnProductCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Update the AddOn product in SAP BTP ABAP Environment system of Landscape Portal",
		Long:  `This step describes the AddOn product update in SAP BTP ABAP Environment system of Landscape Portal`,
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
			log.RegisterSecret(stepConfig.LandscapePortalAPIServiceKey)

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
			abapLandscapePortalUpdateAddOnProduct(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapLandscapePortalUpdateAddOnProductFlags(createAbapLandscapePortalUpdateAddOnProductCmd, &stepConfig)
	return createAbapLandscapePortalUpdateAddOnProductCmd
}

func addAbapLandscapePortalUpdateAddOnProductFlags(cmd *cobra.Command, stepConfig *abapLandscapePortalUpdateAddOnProductOptions) {
	cmd.Flags().StringVar(&stepConfig.LandscapePortalAPIServiceKey, "landscapePortalAPIServiceKey", os.Getenv("PIPER_landscapePortalAPIServiceKey"), "Service key JSON string to access the Landscape Portal Access API")
	cmd.Flags().StringVar(&stepConfig.AbapSystemNumber, "abapSystemNumber", os.Getenv("PIPER_abapSystemNumber"), "System Number of the abap integration test system")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptorFileName, "addonDescriptorFileName", `addon.yml`, "File name of the YAML file which describes the Product Version and corresponding Software Component Versions")

	cmd.MarkFlagRequired("landscapePortalAPIServiceKey")
	cmd.MarkFlagRequired("abapSystemNumber")
	cmd.MarkFlagRequired("addonDescriptorFileName")
}

// retrieve step metadata
func abapLandscapePortalUpdateAddOnProductMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "abapLandscapePortalUpdateAddOnProduct",
			Aliases:     []config.Alias{},
			Description: "Update the AddOn product in SAP BTP ABAP Environment system of Landscape Portal",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "landscapePortalAPICredentialsId", Description: "Jenkins secret text credential ID containing the service key to access the Landscape Portal Access API", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "landscapePortalAPIServiceKey",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "landscapePortalAPICredentialsId",
								Param: "landscapePortalAPIServiceKey",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_landscapePortalAPIServiceKey"),
					},
					{
						Name:        "abapSystemNumber",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_abapSystemNumber"),
					},
					{
						Name:        "addonDescriptorFileName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `addon.yml`,
					},
				},
			},
		},
	}
	return theMetaData
}
