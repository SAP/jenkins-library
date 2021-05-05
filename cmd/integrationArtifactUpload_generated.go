// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type integrationArtifactUploadOptions struct {
	Username               string `json:"username,omitempty"`
	Password               string `json:"password,omitempty"`
	IntegrationFlowID      string `json:"integrationFlowId,omitempty"`
	IntegrationFlowVersion string `json:"integrationFlowVersion,omitempty"`
	IntegrationFlowName    string `json:"integrationFlowName,omitempty"`
	PackageID              string `json:"packageId,omitempty"`
	Host                   string `json:"host,omitempty"`
	OAuthTokenProviderURL  string `json:"oAuthTokenProviderUrl,omitempty"`
	FilePath               string `json:"filePath,omitempty"`
}

// IntegrationArtifactUploadCommand Upload or Update an integration flow designtime artifact
func IntegrationArtifactUploadCommand() *cobra.Command {
	const STEP_NAME = "integrationArtifactUpload"

	metadata := integrationArtifactUploadMetadata()
	var stepConfig integrationArtifactUploadOptions
	var startTime time.Time

	var createIntegrationArtifactUploadCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Upload or Update an integration flow designtime artifact",
		Long:  `With this step you can either upload or update a integration flow designtime artifact using the OData API. Learn more about the SAP Cloud Integration remote API for updating an integration flow artifact [here](https://help.sap.com/viewer/368c481cd6954bdfa5d0435479fd4eaf/Cloud/en-US/d1679a80543f46509a7329243b595bdb.html).`,
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
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
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
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			integrationArtifactUpload(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addIntegrationArtifactUploadFlags(createIntegrationArtifactUploadCmd, &stepConfig)
	return createIntegrationArtifactUploadCmd
}

func addIntegrationArtifactUploadFlags(cmd *cobra.Command, stepConfig *integrationArtifactUploadOptions) {
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User to authenticate to the SAP Cloud Platform Integration Service")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password to authenticate to the SAP Cloud Platform Integration Service")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowID, "integrationFlowId", os.Getenv("PIPER_integrationFlowId"), "Specifies the ID of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowVersion, "integrationFlowVersion", os.Getenv("PIPER_integrationFlowVersion"), "Specifies the version of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowName, "integrationFlowName", os.Getenv("PIPER_integrationFlowName"), "Specifies the Name of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.PackageID, "packageId", os.Getenv("PIPER_packageId"), "Specifies the ID of the Integration Package")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Specifies the protocol and host address, including the port. Please provide in the format `<protocol>://<host>:<port>`. Supported protocols are `http` and `https`.")
	cmd.Flags().StringVar(&stepConfig.OAuthTokenProviderURL, "oAuthTokenProviderUrl", os.Getenv("PIPER_oAuthTokenProviderUrl"), "Specifies the oAuth Provider protocol and host address, including the port. Please provide in the format `<protocol>://<host>:<port>`. Supported protocols are `http` and `https`.")
	cmd.Flags().StringVar(&stepConfig.FilePath, "filePath", os.Getenv("PIPER_filePath"), "Specifies integration artifact relative file path.")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("integrationFlowId")
	cmd.MarkFlagRequired("integrationFlowVersion")
	cmd.MarkFlagRequired("integrationFlowName")
	cmd.MarkFlagRequired("packageId")
	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("oAuthTokenProviderUrl")
	cmd.MarkFlagRequired("filePath")
}

// retrieve step metadata
func integrationArtifactUploadMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "integrationArtifactUpload",
			Aliases:     []config.Alias{},
			Description: "Upload or Update an integration flow designtime artifact",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cpiCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cpiCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:        "integrationFlowId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "integrationFlowVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "integrationFlowName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "packageId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "oAuthTokenProviderUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "filePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
