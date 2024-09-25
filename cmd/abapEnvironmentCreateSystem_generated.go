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
	"go.opentelemetry.io/otel/attribute"
)

type abapEnvironmentCreateSystemOptions struct {
	CfAPIEndpoint                  string `json:"cfApiEndpoint,omitempty"`
	Username                       string `json:"username,omitempty"`
	Password                       string `json:"password,omitempty"`
	CfOrg                          string `json:"cfOrg,omitempty"`
	CfSpace                        string `json:"cfSpace,omitempty"`
	CfService                      string `json:"cfService,omitempty"`
	CfServicePlan                  string `json:"cfServicePlan,omitempty"`
	CfServiceInstance              string `json:"cfServiceInstance,omitempty"`
	ServiceManifest                string `json:"serviceManifest,omitempty"`
	AbapSystemAdminEmail           string `json:"abapSystemAdminEmail,omitempty"`
	AbapSystemDescription          string `json:"abapSystemDescription,omitempty"`
	AbapSystemIsDevelopmentAllowed bool   `json:"abapSystemIsDevelopmentAllowed,omitempty"`
	AbapSystemID                   string `json:"abapSystemID,omitempty"`
	AbapSystemSizeOfPersistence    int    `json:"abapSystemSizeOfPersistence,omitempty"`
	AbapSystemSizeOfRuntime        int    `json:"abapSystemSizeOfRuntime,omitempty"`
	AddonDescriptorFileName        string `json:"addonDescriptorFileName,omitempty"`
	IncludeAddon                   bool   `json:"includeAddon,omitempty"`
}

// AbapEnvironmentCreateSystemCommand Creates a SAP BTP ABAP Environment system (aka Steampunk system)
func AbapEnvironmentCreateSystemCommand() *cobra.Command {
	const STEP_NAME = "abapEnvironmentCreateSystem"

	metadata := abapEnvironmentCreateSystemMetadata()
	var stepConfig abapEnvironmentCreateSystemOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createAbapEnvironmentCreateSystemCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Creates a SAP BTP ABAP Environment system (aka Steampunk system)",
		Long:  `This step creates a SAP BTP ABAP Environment system (aka Steampunk system) via the cloud foundry command line interface (cf CLI). This can be done by providing a service manifest as a configuration file (parameter ` + "`" + `serviceManifest` + "`" + `) or by passing the configuration values directly via the other parameters of this step.`,
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
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

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
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := cmd.Root().Context()
			tracer := telemetry.GetTracer(ctx)
			_, span := tracer.Start(ctx, "piper.step.run")
			span.SetAttributes(attribute.String("piper.step.name", STEP_NAME))

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				defer span.End()
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			abapEnvironmentCreateSystem(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapEnvironmentCreateSystemFlags(createAbapEnvironmentCreateSystemCmd, &stepConfig)
	return createAbapEnvironmentCreateSystemCmd
}

func addAbapEnvironmentCreateSystemFlags(cmd *cobra.Command, stepConfig *abapEnvironmentCreateSystemOptions) {
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", `https://api.cf.eu10.hana.ondemand.com`, "Cloud Foundry API endpoint")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User or E-Mail for CF")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for Cloud Foundry User")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry org")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "Cloud Foundry Space")
	cmd.Flags().StringVar(&stepConfig.CfService, "cfService", os.Getenv("PIPER_cfService"), "Parameter for Cloud Foundry Service to be used for creating Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServicePlan, "cfServicePlan", os.Getenv("PIPER_cfServicePlan"), "Parameter for Cloud Foundry Service Plan to be used when creating a Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServiceInstance, "cfServiceInstance", os.Getenv("PIPER_cfServiceInstance"), "Parameter for naming the Service Instance when creating a Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.ServiceManifest, "serviceManifest", os.Getenv("PIPER_serviceManifest"), "Path to Cloud Foundry Service Manifest in YAML format for multiple service creations that are being passed to a Create-Service-Push Cloud Foundry cli plugin")
	cmd.Flags().StringVar(&stepConfig.AbapSystemAdminEmail, "abapSystemAdminEmail", os.Getenv("PIPER_abapSystemAdminEmail"), "Admin E-Mail address for the initial administrator of the system")
	cmd.Flags().StringVar(&stepConfig.AbapSystemDescription, "abapSystemDescription", `Test system created by an automated pipeline`, "Description for the ABAP Environment system")
	cmd.Flags().BoolVar(&stepConfig.AbapSystemIsDevelopmentAllowed, "abapSystemIsDevelopmentAllowed", true, "This parameter determines, if development is allowed on the system")
	cmd.Flags().StringVar(&stepConfig.AbapSystemID, "abapSystemID", `H02`, "The three character name of the system - maps to 'sapSystemName'")
	cmd.Flags().IntVar(&stepConfig.AbapSystemSizeOfPersistence, "abapSystemSizeOfPersistence", 0, "The size of the persistence")
	cmd.Flags().IntVar(&stepConfig.AbapSystemSizeOfRuntime, "abapSystemSizeOfRuntime", 0, "The size of the runtime")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptorFileName, "addonDescriptorFileName", os.Getenv("PIPER_addonDescriptorFileName"), "The file name of the addonDescriptor")
	cmd.Flags().BoolVar(&stepConfig.IncludeAddon, "includeAddon", false, "Must be set to true to install the addon provided via 'addonDescriptorFileName'")

	cmd.MarkFlagRequired("cfApiEndpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("cfOrg")
	cmd.MarkFlagRequired("cfSpace")
}

// retrieve step metadata
func abapEnvironmentCreateSystemMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "abapEnvironmentCreateSystem",
			Aliases:     []config.Alias{},
			Description: "Creates a SAP BTP ABAP Environment system (aka Steampunk system)",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "cfCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing user and password to authenticate to the Cloud Foundry API.", Type: "jenkins", Aliases: []config.Alias{{Name: "cloudFoundry/credentialsId", Deprecated: false}}},
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

							{
								Name:    "cloudfoundryVaultSecretName",
								Type:    "vaultSecret",
								Default: "cloudfoundry-$(org)-$(space)",
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

							{
								Name:    "cloudfoundryVaultSecretName",
								Type:    "vaultSecret",
								Default: "cloudfoundry-$(org)-$(space)",
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
					{
						Name:        "cfService",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/service"}},
						Default:     os.Getenv("PIPER_cfService"),
					},
					{
						Name:        "cfServicePlan",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/servicePlan"}},
						Default:     os.Getenv("PIPER_cfServicePlan"),
					},
					{
						Name:        "cfServiceInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceInstance"}},
						Default:     os.Getenv("PIPER_cfServiceInstance"),
					},
					{
						Name:        "serviceManifest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceManifest"}, {Name: "cfServiceManifest"}},
						Default:     os.Getenv("PIPER_serviceManifest"),
					},
					{
						Name:        "abapSystemAdminEmail",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_abapSystemAdminEmail"),
					},
					{
						Name:        "abapSystemDescription",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `Test system created by an automated pipeline`,
					},
					{
						Name:        "abapSystemIsDevelopmentAllowed",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "abapSystemID",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `H02`,
					},
					{
						Name:        "abapSystemSizeOfPersistence",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     0,
					},
					{
						Name:        "abapSystemSizeOfRuntime",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     0,
					},
					{
						Name:        "addonDescriptorFileName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_addonDescriptorFileName"),
					},
					{
						Name:        "includeAddon",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
				},
			},
			Containers: []config.Container{
				{Name: "cf", Image: "ppiper/cf-cli:v12"},
			},
		},
	}
	return theMetaData
}
