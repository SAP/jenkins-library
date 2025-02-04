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

type abapAddonAssemblyKitReserveNextPackagesOptions struct {
	AbapAddonAssemblyKitCertificateFile string `json:"abapAddonAssemblyKitCertificateFile,omitempty"`
	AbapAddonAssemblyKitCertificatePass string `json:"abapAddonAssemblyKitCertificatePass,omitempty"`
	AbapAddonAssemblyKitEndpoint        string `json:"abapAddonAssemblyKitEndpoint,omitempty"`
	Username                            string `json:"username,omitempty"`
	Password                            string `json:"password,omitempty"`
	AddonDescriptor                     string `json:"addonDescriptor,omitempty"`
	MaxRuntimeInMinutes                 int    `json:"maxRuntimeInMinutes,omitempty"`
	PollingIntervalInSeconds            int    `json:"pollingIntervalInSeconds,omitempty"`
	AbapAddonAssemblyKitOriginHash      string `json:"abapAddonAssemblyKitOriginHash,omitempty"`
}

type abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment struct {
	abap struct {
		addonDescriptor string
	}
}

func (p *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "abap", name: "addonDescriptor", value: p.abap.addonDescriptor},
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

// AbapAddonAssemblyKitReserveNextPackagesCommand This step determines the ABAP delivery packages (name and type), which are needed to deliver Software Component Versions.
func AbapAddonAssemblyKitReserveNextPackagesCommand() *cobra.Command {
	const STEP_NAME = "abapAddonAssemblyKitReserveNextPackages"

	metadata := abapAddonAssemblyKitReserveNextPackagesMetadata()
	var stepConfig abapAddonAssemblyKitReserveNextPackagesOptions
	var startTime time.Time
	var commonPipelineEnvironment abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createAbapAddonAssemblyKitReserveNextPackagesCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step determines the ABAP delivery packages (name and type), which are needed to deliver Software Component Versions.",
		Long: `This step takes the list of Software Component Versions from the addonDescriptor in the commonPipelineEnvironment and determines the ABAP delivery packages.
If a package does not exist yet in the package registry, it is created there. The response contains detail information for this package and a package status, which determines the next actions:
"P": Package was created in the registry; production can be started / continued
"R": Package exists and is already released; production is not needed and will be skipped.
The step waits until the status "P" or "R" is achieved.
The name, type and namespace of each package is written back to the addonDescriptor in the commonPipelineEnvironment.
<br />
For logon you can either provide a credential with basic authorization (username and password) or two secret text credentials containing the technical s-users certificate (see note [2805811](https://me.sap.com/notes/2805811) for download) as base64 encoded string and the password to decrypt the file
<br />
For Terminology refer to the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).`,
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
			log.RegisterSecret(stepConfig.AbapAddonAssemblyKitCertificateFile)
			log.RegisterSecret(stepConfig.AbapAddonAssemblyKitCertificatePass)
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)
			log.RegisterSecret(stepConfig.AbapAddonAssemblyKitOriginHash)

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
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			abapAddonAssemblyKitReserveNextPackages(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapAddonAssemblyKitReserveNextPackagesFlags(createAbapAddonAssemblyKitReserveNextPackagesCmd, &stepConfig)
	return createAbapAddonAssemblyKitReserveNextPackagesCmd
}

func addAbapAddonAssemblyKitReserveNextPackagesFlags(cmd *cobra.Command, stepConfig *abapAddonAssemblyKitReserveNextPackagesOptions) {
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitCertificateFile, "abapAddonAssemblyKitCertificateFile", os.Getenv("PIPER_abapAddonAssemblyKitCertificateFile"), "base64 encoded certificate pfx file (PKCS12 format) see note [2805811](https://me.sap.com/notes/2805811)")
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitCertificatePass, "abapAddonAssemblyKitCertificatePass", os.Getenv("PIPER_abapAddonAssemblyKitCertificatePass"), "password to decrypt the certificate file")
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitEndpoint, "abapAddonAssemblyKitEndpoint", `https://apps.support.sap.com`, "Base URL to the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User for the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptor, "addonDescriptor", os.Getenv("PIPER_addonDescriptor"), "Structure in the commonPipelineEnvironment containing information about the Product Version and corresponding Software Component Versions")
	cmd.Flags().IntVar(&stepConfig.MaxRuntimeInMinutes, "maxRuntimeInMinutes", 5, "Maximum runtime for status polling in minutes")
	cmd.Flags().IntVar(&stepConfig.PollingIntervalInSeconds, "pollingIntervalInSeconds", 30, "Wait time in seconds between polling calls")
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitOriginHash, "abapAddonAssemblyKitOriginHash", os.Getenv("PIPER_abapAddonAssemblyKitOriginHash"), "Origin Hash for restricted AAKaaS scenarios")

	cmd.MarkFlagRequired("abapAddonAssemblyKitEndpoint")
	cmd.MarkFlagRequired("addonDescriptor")
}

// retrieve step metadata
func abapAddonAssemblyKitReserveNextPackagesMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "abapAddonAssemblyKitReserveNextPackages",
			Aliases:     []config.Alias{},
			Description: "This step determines the ABAP delivery packages (name and type), which are needed to deliver Software Component Versions.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapAddonAssemblyKitCredentialsId", Description: "Credential stored in Jenkins for the Addon Assembly Kit as a Service (AAKaaS) system", Type: "jenkins"},
					{Name: "abapAddonAssemblyKitCertificateFileCredentialsId", Description: "Jenkins secret text credential ID containing the base64 encoded certificate pfx file (PKCS12 format) see note [2805811](https://me.sap.com/notes/2805811)", Type: "jenkins"},
					{Name: "abapAddonAssemblyKitCertificatePassCredentialsId", Description: "Jenkins secret text credential ID containing the password to decrypt the certificate file stored in abapAddonAssemblyKitCertificateFileCredentialsId", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "abapAddonAssemblyKitCertificateFile",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapAddonAssemblyKitCertificateFileCredentialsId",
								Param: "abapAddonAssemblyKitCertificateFile",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_abapAddonAssemblyKitCertificateFile"),
					},
					{
						Name: "abapAddonAssemblyKitCertificatePass",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapAddonAssemblyKitCertificatePassCredentialsId",
								Param: "abapAddonAssemblyKitCertificatePass",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_abapAddonAssemblyKitCertificatePass"),
					},
					{
						Name:        "abapAddonAssemblyKitEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `https://apps.support.sap.com`,
					},
					{
						Name:        "username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_username"),
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_password"),
					},
					{
						Name: "addonDescriptor",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "abap/addonDescriptor",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_addonDescriptor"),
					},
					{
						Name:        "maxRuntimeInMinutes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     5,
					},
					{
						Name:        "pollingIntervalInSeconds",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     30,
					},
					{
						Name:        "abapAddonAssemblyKitOriginHash",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_abapAddonAssemblyKitOriginHash"),
					},
				},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "abap/addonDescriptor"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
