// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type transportRequestUploadRFCOptions struct {
	Endpoint                   string `json:"endpoint,omitempty"`
	Instance                   string `json:"instance,omitempty"`
	Username                   string `json:"username,omitempty"`
	Password                   string `json:"password,omitempty"`
	Client                     string `json:"client,omitempty"`
	ApplicationName            string `json:"applicationName,omitempty"`
	ApplicationDescription     string `json:"applicationDescription,omitempty"`
	AbapPackage                string `json:"abapPackage,omitempty"`
	ApplicationURL             string `json:"applicationUrl,omitempty"`
	CodePage                   string `json:"codePage,omitempty"`
	AcceptUnixStyleLineEndings bool   `json:"acceptUnixStyleLineEndings,omitempty"`
	FailUploadOnWarning        bool   `json:"failUploadOnWarning,omitempty"`
	TransportRequestID         string `json:"transportRequestId,omitempty"`
}

type transportRequestUploadRFCCommonPipelineEnvironment struct {
	custom struct {
		transportRequestID string
	}
}

func (p *transportRequestUploadRFCCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "transportRequestId", value: p.custom.transportRequestID},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Piper environment")
	}
}

// TransportRequestUploadRFCCommand This step uploads a UI5 application as ZIP file to the ABAP system via RFC connections.
func TransportRequestUploadRFCCommand() *cobra.Command {
	const STEP_NAME = "transportRequestUploadRFC"

	metadata := transportRequestUploadRFCMetadata()
	var stepConfig transportRequestUploadRFCOptions
	var startTime time.Time
	var commonPipelineEnvironment transportRequestUploadRFCCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createTransportRequestUploadRFCCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step uploads a UI5 application as ZIP file to the ABAP system via RFC connections.",
		Long:  `This step uploads a UI5 application as ZIP file to the ABAP system via RFC connections.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, err := os.Getwd()
			if err != nil {
				return err
			}
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err = PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
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
			vaultClient := config.GlobalVaultClient()
			if vaultClient != nil {
				defer vaultClient.MustRevokeToken()
			}

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.LogStepTelemetryData()
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
			telemetryClient.Initialize(STEP_NAME)
			transportRequestUploadRFC(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTransportRequestUploadRFCFlags(createTransportRequestUploadRFCCmd, &stepConfig)
	return createTransportRequestUploadRFCCmd
}

func addTransportRequestUploadRFCFlags(cmd *cobra.Command, stepConfig *transportRequestUploadRFCOptions) {
	cmd.Flags().StringVar(&stepConfig.Endpoint, "endpoint", os.Getenv("PIPER_endpoint"), "Service endpoint, Application server URL")
	cmd.Flags().StringVar(&stepConfig.Instance, "instance", os.Getenv("PIPER_instance"), "AS ABAP instance number")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "Service user for uploading to the ABAP system via RFC")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Service user password for uploading to the ABAP system via RFC")
	cmd.Flags().StringVar(&stepConfig.Client, "client", os.Getenv("PIPER_client"), "AS ABAP client number")
	cmd.Flags().StringVar(&stepConfig.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "Name of the UI5 application")
	cmd.Flags().StringVar(&stepConfig.ApplicationDescription, "applicationDescription", os.Getenv("PIPER_applicationDescription"), "Description of the UI5 application")
	cmd.Flags().StringVar(&stepConfig.AbapPackage, "abapPackage", os.Getenv("PIPER_abapPackage"), "ABAP package name of the UI5 application")
	cmd.Flags().StringVar(&stepConfig.ApplicationURL, "applicationUrl", os.Getenv("PIPER_applicationUrl"), "URL of the UI5 application package to upload to the ABAP system via RFC")
	cmd.Flags().StringVar(&stepConfig.CodePage, "codePage", `UTF-8`, "Code page")
	cmd.Flags().BoolVar(&stepConfig.AcceptUnixStyleLineEndings, "acceptUnixStyleLineEndings", true, "If unix style line endings should be accepted")
	cmd.Flags().BoolVar(&stepConfig.FailUploadOnWarning, "failUploadOnWarning", true, "If the upload should fail in case the log contains warnings")
	cmd.Flags().StringVar(&stepConfig.TransportRequestID, "transportRequestId", os.Getenv("PIPER_transportRequestId"), "ID of the transport request to which the UI5 application is uploaded")

	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("applicationName")
	cmd.MarkFlagRequired("abapPackage")
	cmd.MarkFlagRequired("applicationUrl")
	cmd.MarkFlagRequired("transportRequestId")
}

// retrieve step metadata
func transportRequestUploadRFCMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "transportRequestUploadRFC",
			Aliases:     []config.Alias{{Name: "transportRequestUploadFile", Deprecated: false}},
			Description: "This step uploads a UI5 application as ZIP file to the ABAP system via RFC connections.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "uploadCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing user and password to authenticate against the ABAP system", Type: "jenkins", Aliases: []config.Alias{{Name: "changeManagement/credentialsId", Deprecated: false}}},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "endpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "changeManagement/endpoint"}},
						Default:     os.Getenv("PIPER_endpoint"),
					},
					{
						Name:        "instance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/instance"}, {Name: "changeManagement/rfc/developmentInstance"}},
						Default:     os.Getenv("PIPER_instance"),
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "uploadCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_username"),
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "uploadCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_password"),
					},
					{
						Name:        "client",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/client"}, {Name: "changeManagement/rfc/developmentClient"}},
						Default:     os.Getenv("PIPER_client"),
					},
					{
						Name:        "applicationName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationName"),
					},
					{
						Name:        "applicationDescription",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationDescription"),
					},
					{
						Name:        "abapPackage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_abapPackage"),
					},
					{
						Name:        "applicationUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationUrl"),
					},
					{
						Name:        "codePage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `UTF-8`,
					},
					{
						Name:        "acceptUnixStyleLineEndings",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "failUploadOnWarning",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "failOnWarning"}},
						Default:     true,
					},
					{
						Name: "transportRequestId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/transportRequestId",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_transportRequestId"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "rfcclient", Image: "rfc-client"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "custom/transportRequestId"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
