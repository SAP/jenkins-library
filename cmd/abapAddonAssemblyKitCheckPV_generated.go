// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type abapAddonAssemblyKitCheckPVOptions struct {
	AbapAddonAssemblyKitCertificateFile string `json:"abapAddonAssemblyKitCertificateFile,omitempty"`
	AbapAddonAssemblyKitCertificatePass string `json:"abapAddonAssemblyKitCertificatePass,omitempty"`
	AbapAddonAssemblyKitEndpoint        string `json:"abapAddonAssemblyKitEndpoint,omitempty"`
	Username                            string `json:"username,omitempty"`
	Password                            string `json:"password,omitempty"`
	AddonDescriptorFileName             string `json:"addonDescriptorFileName,omitempty"`
	AddonDescriptor                     string `json:"addonDescriptor,omitempty"`
}

type abapAddonAssemblyKitCheckPVCommonPipelineEnvironment struct {
	abap struct {
		addonDescriptor string
	}
}

func (p *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment) persist(path, resourceName string) {
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

// AbapAddonAssemblyKitCheckPVCommand This step checks the validity of a Addon Product Version.
func AbapAddonAssemblyKitCheckPVCommand() *cobra.Command {
	const STEP_NAME = "abapAddonAssemblyKitCheckPV"

	metadata := abapAddonAssemblyKitCheckPVMetadata()
	var stepConfig abapAddonAssemblyKitCheckPVOptions
	var startTime time.Time
	var commonPipelineEnvironment abapAddonAssemblyKitCheckPVCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createAbapAddonAssemblyKitCheckPVCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step checks the validity of a Addon Product Version.",
		Long: `This step checks by calling AAKaaS whether the Addon Product Version in the addonDescriptor configuration file specified via addonDescriptorFileName (e.g. addon.yml) does exist or is a valid successor of an existing Product Version.
It resolves the dotted version string into version, support package stack level and patch level and writes it to the addonDescriptor structure in the Piper commonPipelineEnvironment for usage of subsequent pipeline steps.
<br />
For logon you can either provide a credential with basic authorization (username and password) or two credentials secret text credentials containing the technical s-users certificate (see note [2805811](https://me.sap.com/notes/2805811) for download) as base64 encoded string and the password to decrypt the file
<br />
For Terminology refer to the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).`,
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
			log.RegisterSecret(stepConfig.AbapAddonAssemblyKitCertificateFile)
			log.RegisterSecret(stepConfig.AbapAddonAssemblyKitCertificatePass)
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
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			abapAddonAssemblyKitCheckPV(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapAddonAssemblyKitCheckPVFlags(createAbapAddonAssemblyKitCheckPVCmd, &stepConfig)
	return createAbapAddonAssemblyKitCheckPVCmd
}

func addAbapAddonAssemblyKitCheckPVFlags(cmd *cobra.Command, stepConfig *abapAddonAssemblyKitCheckPVOptions) {
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitCertificateFile, "abapAddonAssemblyKitCertificateFile", os.Getenv("PIPER_abapAddonAssemblyKitCertificateFile"), "base64 encoded certificate pfx file (PKCS12 format) see note [2805811](https://me.sap.com/notes/2805811)")
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitCertificatePass, "abapAddonAssemblyKitCertificatePass", os.Getenv("PIPER_abapAddonAssemblyKitCertificatePass"), "password to decrypt the certificate file")
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitEndpoint, "abapAddonAssemblyKitEndpoint", `https://apps.support.sap.com`, "Base URL to the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User for the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptorFileName, "addonDescriptorFileName", `addon.yml`, "File name of the YAML file which describes the Product Version and corresponding Software Component Versions")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptor, "addonDescriptor", os.Getenv("PIPER_addonDescriptor"), "Structure in the commonPipelineEnvironment containing information about the Product Version and corresponding Software Component Versions")

	cmd.MarkFlagRequired("abapAddonAssemblyKitEndpoint")
	cmd.MarkFlagRequired("addonDescriptorFileName")
}

// retrieve step metadata
func abapAddonAssemblyKitCheckPVMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "abapAddonAssemblyKitCheckPV",
			Aliases:     []config.Alias{},
			Description: "This step checks the validity of a Addon Product Version.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapAddonAssemblyKitCredentialsId", Description: "CredentialsId stored in Jenkins for the Addon Assembly Kit as a Service (AAKaaS) system", Type: "jenkins"},
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
						Name:        "addonDescriptorFileName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `addon.yml`,
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
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_addonDescriptor"),
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
