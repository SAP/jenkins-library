// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/cmd/piper"
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type cloudFoundryCreateSpaceOptions struct {
	CfAPIEndpoint string `json:"cfApiEndpoint,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	CfOrg         string `json:"cfOrg,omitempty"`
	CfSpace       string `json:"cfSpace,omitempty"`
}

// CloudFoundryCreateSpaceCommand Creates a user defined space in Cloud Foundry
func CloudFoundryCreateSpaceCommand() *cobra.Command {
	const STEP_NAME = "cloudFoundryCreateSpace"

	metadata := CloudFoundryCreateSpaceMetadata()
	var stepConfig cloudFoundryCreateSpaceOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}
	var GeneralConfig piper.GeneralConfigOptions

	var createCloudFoundryCreateSpaceCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Creates a user defined space in Cloud Foundry",
		Long: `Creates a cf space in Cloud Foundry
Mandatory:
* Cloud Foundry API endpoint, Organization, name of the Cf space to be created`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			var err error
			GeneralConfig, err = piper.CreateGeneralConfigCopyFromFlags(cmd)
			log.SetVerbose(GeneralConfig.Verbose)

			//GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err = GeneralConfig.PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
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
				stepTelemetryData.PiperCommitHash = GeneralConfig.GitCommit
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
			cloudFoundryCreateSpace(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addCloudFoundryCreateSpaceFlags(createCloudFoundryCreateSpaceCmd, &stepConfig)
	return createCloudFoundryCreateSpaceCmd
}

func addCloudFoundryCreateSpaceFlags(cmd *cobra.Command, stepConfig *cloudFoundryCreateSpaceOptions) {
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", `https://api.cf.eu10.hana.ondemand.com`, "Cloud Foundry API endpoint")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User or E-Mail for CF")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for Cloud Foundry User")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry org")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "The name of the Cloud Foundry Space to be created")

	cmd.MarkFlagRequired("cfApiEndpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("cfOrg")
	cmd.MarkFlagRequired("cfSpace")
}

// retrieve step metadata
func CloudFoundryCreateSpaceMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "cloudFoundryCreateSpace",
			Aliases:     []config.Alias{},
			Description: "Creates a user defined space in Cloud Foundry",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "cfCredentialsId", Description: "Jenkins credentials ID containing user and password to authenticate to the Cloud Foundry API", Type: "jenkins", Aliases: []config.Alias{{Name: "cloudFoundry/credentialsId", Deprecated: false}}},
				},
				Resources: []config.StepResources{
					{Name: "deployDescriptor", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "cfApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
						Default:     `https://api.cf.eu10.hana.ondemand.com`,
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cfCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_username"),
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cfCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_password"),
					},
					{
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/org"}},
						Default:     os.Getenv("PIPER_cfOrg"),
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/space"}},
						Default:     os.Getenv("PIPER_cfSpace"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "cf", Image: "ppiper/cf-cli:latest"},
			},
		},
	}
	return theMetaData
}
