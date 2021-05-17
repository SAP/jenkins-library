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

type transportRequestUploadCTSOptions struct {
	Description            string   `json:"description,omitempty"`
	Endpoint               string   `json:"endpoint,omitempty"`
	Client                 string   `json:"client,omitempty"`
	Username               string   `json:"username,omitempty"`
	Password               string   `json:"password,omitempty"`
	ApplicationName        string   `json:"applicationName,omitempty"`
	AbapPackage            string   `json:"abapPackage,omitempty"`
	OsDeployUser           string   `json:"osDeployUser,omitempty"`
	DeployConfigFile       string   `json:"deployConfigFile,omitempty"`
	TransportRequestID     string   `json:"transportRequestId,omitempty"`
	DeployToolDependencies []string `json:"deployToolDependencies,omitempty"`
	NpmInstallOpts         []string `json:"npmInstallOpts,omitempty"`
}

// TransportRequestUploadCTSCommand Uploads content to a transport request
func TransportRequestUploadCTSCommand() *cobra.Command {
	const STEP_NAME = "transportRequestUploadCTS"

	metadata := transportRequestUploadCTSMetadata()
	var stepConfig transportRequestUploadCTSOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createTransportRequestUploadCTSCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Uploads content to a transport request",
		Long:  `Uploads content to a transport request.`,
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
			transportRequestUploadCTS(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTransportRequestUploadCTSFlags(createTransportRequestUploadCTSCmd, &stepConfig)
	return createTransportRequestUploadCTSCmd
}

func addTransportRequestUploadCTSFlags(cmd *cobra.Command, stepConfig *transportRequestUploadCTSOptions) {
	cmd.Flags().StringVar(&stepConfig.Description, "description", `Deployed with Piper based on SAP Fiori tools`, "The description of the application. the desription is only taken into account for a new upload. In case of an update the description will not be updated.")
	cmd.Flags().StringVar(&stepConfig.Endpoint, "endpoint", os.Getenv("PIPER_endpoint"), "The service endpoint")
	cmd.Flags().StringVar(&stepConfig.Client, "client", os.Getenv("PIPER_client"), "The ABAP client")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "The deploy user")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "The password for the deploy user")
	cmd.Flags().StringVar(&stepConfig.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "The name of the application.")
	cmd.Flags().StringVar(&stepConfig.AbapPackage, "abapPackage", os.Getenv("PIPER_abapPackage"), "The ABAP package name of your application")
	cmd.Flags().StringVar(&stepConfig.OsDeployUser, "osDeployUser", `node`, "By default we use a standard node docker image and prepare some fiori related packages before performing the deployment. For that we need to launch the image with root privileges. After that, before actually performing the deployment we swith to a non root user. This user can be specified here.")
	cmd.Flags().StringVar(&stepConfig.DeployConfigFile, "deployConfigFile", `ui5-deploy.yaml`, "The ABAP package name of your application")
	cmd.Flags().StringVar(&stepConfig.TransportRequestID, "transportRequestId", os.Getenv("PIPER_transportRequestId"), "The id of the transport request to upload the file. This parameter is only taken into account when provided via signature to the step.")
	cmd.Flags().StringSliceVar(&stepConfig.DeployToolDependencies, "deployToolDependencies", []string{}, "By default we use a standard node docker iamge and prepare some fiori related packages performing the deployment. The additional dependencies can be provided here. In case you use an already prepared docker image which contains the required dependencies, the empty list can be provide here. Caused hereby installing additional dependencies will be skipped.")
	cmd.Flags().StringSliceVar(&stepConfig.NpmInstallOpts, "npmInstallOpts", []string{}, "A list containing additional options for the npm install call. `-g`, `--global` is always assumed. Can be used for e.g. providing custom registries (`--registry https://your.registry.com`) or for providing the verbose flag (`--verbose`) for troubleshooting.")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("applicationName")
	cmd.MarkFlagRequired("abapPackage")
	cmd.MarkFlagRequired("transportRequestId")
}

// retrieve step metadata
func transportRequestUploadCTSMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "transportRequestUploadCTS",
			Aliases:     []config.Alias{{Name: "transportRequestUploadFile", Deprecated: false}},
			Description: "Uploads content to a transport request",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "description",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "endpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/endpoint"}},
					},
					{
						Name:        "client",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/client"}},
					},
					{
						Name:        "username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "applicationName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "abapPackage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "osDeployUser",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/cts/osDeployUser"}},
					},
					{
						Name:        "deployConfigFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/cts/deployConfigFile"}, {Name: "cts/deployConfigFile"}},
					},
					{
						Name:        "transportRequestId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "deployToolDependencies",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/cts/deployToolDependencies"}},
					},
					{
						Name:        "npmInstallOpts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/cts/deployToolDependencies"}},
					},
				},
			},
		},
	}
	return theMetaData
}
