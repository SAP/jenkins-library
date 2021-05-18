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

type vaultRotateSecretIdOptions struct {
	SecretStore                          string `json:"secretStore,omitempty"`
	JenkinsURL                           string `json:"jenkinsUrl,omitempty"`
	JenkinsCredentialDomain              string `json:"jenkinsCredentialDomain,omitempty"`
	JenkinsUsername                      string `json:"jenkinsUsername,omitempty"`
	JenkinsToken                         string `json:"jenkinsToken,omitempty"`
	VaultAppRoleSecretTokenCredentialsID string `json:"vaultAppRoleSecretTokenCredentialsId,omitempty"`
	VaultServerURL                       string `json:"vaultServerUrl,omitempty"`
	VaultNamespace                       string `json:"vaultNamespace,omitempty"`
	DaysBeforeExpiry                     int    `json:"daysBeforeExpiry,omitempty"`
}

// VaultRotateSecretIdCommand Rotate vault AppRole Secret ID
func VaultRotateSecretIdCommand() *cobra.Command {
	const STEP_NAME = "vaultRotateSecretId"

	metadata := vaultRotateSecretIdMetadata()
	var stepConfig vaultRotateSecretIdOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createVaultRotateSecretIdCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Rotate vault AppRole Secret ID",
		Long:  `This step takes the given Vault secret ID and checks whether it needs to be renewed and if so it will update the secret ID in the configured secret store.`,
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
			log.RegisterSecret(stepConfig.JenkinsURL)
			log.RegisterSecret(stepConfig.JenkinsUsername)
			log.RegisterSecret(stepConfig.JenkinsToken)

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
			vaultRotateSecretId(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addVaultRotateSecretIdFlags(createVaultRotateSecretIdCmd, &stepConfig)
	return createVaultRotateSecretIdCmd
}

func addVaultRotateSecretIdFlags(cmd *cobra.Command, stepConfig *vaultRotateSecretIdOptions) {
	cmd.Flags().StringVar(&stepConfig.SecretStore, "secretStore", `jenkins`, "The store to which the secret should be written back to")
	cmd.Flags().StringVar(&stepConfig.JenkinsURL, "jenkinsUrl", os.Getenv("PIPER_jenkinsUrl"), "The jenkins url")
	cmd.Flags().StringVar(&stepConfig.JenkinsCredentialDomain, "jenkinsCredentialDomain", `_`, "The jenkins credential domain which should be used")
	cmd.Flags().StringVar(&stepConfig.JenkinsUsername, "jenkinsUsername", os.Getenv("PIPER_jenkinsUsername"), "The jenkins username")
	cmd.Flags().StringVar(&stepConfig.JenkinsToken, "jenkinsToken", os.Getenv("PIPER_jenkinsToken"), "The jenkins token")
	cmd.Flags().StringVar(&stepConfig.VaultAppRoleSecretTokenCredentialsID, "vaultAppRoleSecretTokenCredentialsId", os.Getenv("PIPER_vaultAppRoleSecretTokenCredentialsId"), "The Jenkins credential ID for the Vault AppRole Secret ID credential")
	cmd.Flags().StringVar(&stepConfig.VaultServerURL, "vaultServerUrl", os.Getenv("PIPER_vaultServerUrl"), "The URL for the Vault server to use")
	cmd.Flags().StringVar(&stepConfig.VaultNamespace, "vaultNamespace", os.Getenv("PIPER_vaultNamespace"), "The vault namespace that should be used (optional)")
	cmd.Flags().IntVar(&stepConfig.DaysBeforeExpiry, "daysBeforeExpiry", 15, "The amount of days before expiry until the secret ID gets rotated")

	cmd.MarkFlagRequired("vaultAppRoleSecretTokenCredentialsId")
	cmd.MarkFlagRequired("vaultServerUrl")
}

// retrieve step metadata
func vaultRotateSecretIdMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "vaultRotateSecretId",
			Aliases:     []config.Alias{},
			Description: "Rotate vault AppRole Secret ID",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "secretStore",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name: "jenkinsUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "",
								Paths: []string{"$(vaultPath)/jenkins", "$(vaultBasePath)/$(vaultPipelineName)/jenkins", "$(vaultBasePath)/GROUP-SECRETS/jenkins"},
								Type:  "vaultSecret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "url"}},
					},
					{
						Name:        "jenkinsCredentialDomain",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name: "jenkinsUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "",
								Paths: []string{"$(vaultPath)/jenkins", "$(vaultBasePath)/$(vaultPipelineName)/jenkins", "$(vaultBasePath)/GROUP-SECRETS/jenkins"},
								Type:  "vaultSecret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "userId"}},
					},
					{
						Name: "jenkinsToken",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "",
								Paths: []string{"$(vaultPath)/jenkins", "$(vaultBasePath)/$(vaultPipelineName)/jenkins", "$(vaultBasePath)/GROUP-SECRETS/jenkins"},
								Type:  "vaultSecret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "token"}},
					},
					{
						Name:        "vaultAppRoleSecretTokenCredentialsId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "vaultServerUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "vaultNamespace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "daysBeforeExpiry",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
