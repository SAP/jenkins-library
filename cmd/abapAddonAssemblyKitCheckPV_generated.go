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
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type abapAddonAssemblyKitCheckPVOptions struct {
	AbapAddonAssemblyKitEndpoint string `json:"abapAddonAssemblyKitEndpoint,omitempty"`
	Username                     string `json:"username,omitempty"`
	Password                     string `json:"password,omitempty"`
	AddonDescriptorFileName      string `json:"addonDescriptorFileName,omitempty"`
	AddonDescriptor              string `json:"addonDescriptor,omitempty"`
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
		log.Entry().Fatal("failed to persist Piper environment")
	}
}

// AbapAddonAssemblyKitCheckPVCommand This step checks the validity of a Addon Product Version.
func AbapAddonAssemblyKitCheckPVCommand() *cobra.Command {
	const STEP_NAME = "abapAddonAssemblyKitCheckPV"

	metadata := abapAddonAssemblyKitCheckPVMetadata()
	var stepConfig abapAddonAssemblyKitCheckPVOptions
	var startTime time.Time
	var commonPipelineEnvironment abapAddonAssemblyKitCheckPVCommonPipelineEnvironment

	var createAbapAddonAssemblyKitCheckPVCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step checks the validity of a Addon Product Version.",
		Long: `This step checks whether the Addon Product Version in the addonDescriptorFileName does exist or is a valid successor of an existing Product Version.
It resolves the dotted version string into version, support package stack level and patch level and writes it to the commonPipelineEnvironment.`,
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
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			abapAddonAssemblyKitCheckPV(stepConfig, &telemetryData, &commonPipelineEnvironment)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapAddonAssemblyKitCheckPVFlags(createAbapAddonAssemblyKitCheckPVCmd, &stepConfig)
	return createAbapAddonAssemblyKitCheckPVCmd
}

func addAbapAddonAssemblyKitCheckPVFlags(cmd *cobra.Command, stepConfig *abapAddonAssemblyKitCheckPVOptions) {
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitEndpoint, "abapAddonAssemblyKitEndpoint", `https://apps.support.sap.com`, "Base URL to the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User for the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptorFileName, "addonDescriptorFileName", `addon.yml`, "File name of the YAML file which describes the Product Version and corresponding Software Component Versions")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptor, "addonDescriptor", os.Getenv("PIPER_addonDescriptor"), "Structure in the commonPipelineEnvironment containing information about the Product Version and corresponding Software Component Versions")

	cmd.MarkFlagRequired("abapAddonAssemblyKitEndpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
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
					{Name: "abapAddonAssemblyKitCredentialsId", Description: "Credential stored in Jenkins for the Addon Assembly Kit as a Service (AAKaaS) system", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "abapAddonAssemblyKitEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Default:     `https://apps.support.sap.com`,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_username"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_password"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "addonDescriptorFileName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{},
						Type:        "string",
						Mandatory:   true,
						Default:     `addon.yml`,
						Aliases:     []config.Alias{},
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
						Default:   os.Getenv("PIPER_addonDescriptor"),
						Aliases:   []config.Alias{},
					},
				},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"Name": "abap/addonDescriptor"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
