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

type abapEnvironmentAssemblePackagesOptions struct {
	CfAPIEndpoint                string `json:"cfApiEndpoint,omitempty"`
	CfOrg                        string `json:"cfOrg,omitempty"`
	CfSpace                      string `json:"cfSpace,omitempty"`
	CfServiceInstance            string `json:"cfServiceInstance,omitempty"`
	CfServiceKeyName             string `json:"cfServiceKeyName,omitempty"`
	Host                         string `json:"host,omitempty"`
	Username                     string `json:"username,omitempty"`
	Password                     string `json:"password,omitempty"`
	AddonDescriptor              string `json:"addonDescriptor,omitempty"`
	MaxRuntimeInMinutes          int    `json:"maxRuntimeInMinutes,omitempty"`
	PollIntervalsInMilliseconds  int    `json:"pollIntervalsInMilliseconds,omitempty"`
	PerformAssemblePackages      bool   `json:"performAssemblePackages,omitempty"`
	PerformRegisterPackages      bool   `json:"performRegisterPackages,omitempty"`
	AbapAddonAssemblyKitEndpoint string `json:"abapAddonAssemblyKitEndpoint,omitempty"`
	AakUsername                  string `json:"aakUsername,omitempty"`
	AakPassword                  string `json:"aakPassword,omitempty"`
}

type abapEnvironmentAssemblePackagesCommonPipelineEnvironment struct {
	abap struct {
		addonDescriptor string
	}
}

func (p *abapEnvironmentAssemblePackagesCommonPipelineEnvironment) persist(path, resourceName string) {
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

// AbapEnvironmentAssemblePackagesCommand Assembly of installation, support package or patch in SAP Cloud Platform ABAP Environment system
func AbapEnvironmentAssemblePackagesCommand() *cobra.Command {
	const STEP_NAME = "abapEnvironmentAssemblePackages"

	metadata := abapEnvironmentAssemblePackagesMetadata()
	var stepConfig abapEnvironmentAssemblePackagesOptions
	var startTime time.Time
	var commonPipelineEnvironment abapEnvironmentAssemblePackagesCommonPipelineEnvironment
	var logCollector *log.CollectorHook

	var createAbapEnvironmentAssemblePackagesCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Assembly of installation, support package or patch in SAP Cloud Platform ABAP Environment system",
		Long: `This step runs the assembly of a list of provided [installations, support packages or patches](https://help.sap.com/viewer/9043aa5d2f834ad385e1cdfdadc06b6f/LATEST/en-US/9a81f55473568c77e10000000a174cb4.html) in SAP Cloud
Platform ABAP Environment system and saves the corresponding [SAR archive](https://launchpad.support.sap.com/#/notes/212876) to the filesystem.`,
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
			log.RegisterSecret(stepConfig.AakUsername)
			log.RegisterSecret(stepConfig.AakPassword)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
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
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
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
			abapEnvironmentAssemblePackages(stepConfig, &telemetryData, &commonPipelineEnvironment)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapEnvironmentAssemblePackagesFlags(createAbapEnvironmentAssemblePackagesCmd, &stepConfig)
	return createAbapEnvironmentAssemblePackagesCmd
}

func addAbapEnvironmentAssemblePackagesFlags(cmd *cobra.Command, stepConfig *abapEnvironmentAssemblePackagesOptions) {
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", os.Getenv("PIPER_cfApiEndpoint"), "Cloud Foundry API endpoint")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry target organization")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "Cloud Foundry target space")
	cmd.Flags().StringVar(&stepConfig.CfServiceInstance, "cfServiceInstance", os.Getenv("PIPER_cfServiceInstance"), "Cloud Foundry Service Instance")
	cmd.Flags().StringVar(&stepConfig.CfServiceKeyName, "cfServiceKeyName", os.Getenv("PIPER_cfServiceKeyName"), "Cloud Foundry Service Key")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Specifies the host address of the SAP Cloud Platform ABAP Environment system")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0582")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for either the Cloud Foundry API or the Communication Arrangement for SAP_COM_0582")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptor, "addonDescriptor", os.Getenv("PIPER_addonDescriptor"), "Structure in the commonPipelineEnvironment containing information about the Product Version and corresponding Software Component Versions")
	cmd.Flags().IntVar(&stepConfig.MaxRuntimeInMinutes, "maxRuntimeInMinutes", 360, "maximal runtime of the step in minutes")
	cmd.Flags().IntVar(&stepConfig.PollIntervalsInMilliseconds, "pollIntervalsInMilliseconds", 60000, "wait time in milliseconds till next status request in the backend system")
	cmd.Flags().BoolVar(&stepConfig.PerformAssemblePackages, "performAssemblePackages", true, "Defines if packages should be assembled in the build system")
	cmd.Flags().BoolVar(&stepConfig.PerformRegisterPackages, "performRegisterPackages", true, "Defines if packages should be registered in AAKaaS")
	cmd.Flags().StringVar(&stepConfig.AbapAddonAssemblyKitEndpoint, "abapAddonAssemblyKitEndpoint", `https://apps.support.sap.com`, "Base URL to the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.AakUsername, "aakUsername", os.Getenv("PIPER_aakUsername"), "User for the Addon Assembly Kit as a Service (AAKaaS) system")
	cmd.Flags().StringVar(&stepConfig.AakPassword, "aakPassword", os.Getenv("PIPER_aakPassword"), "Password for the Addon Assembly Kit as a Service (AAKaaS) system")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("addonDescriptor")
	cmd.MarkFlagRequired("maxRuntimeInMinutes")
	cmd.MarkFlagRequired("pollIntervalsInMilliseconds")
	cmd.MarkFlagRequired("abapAddonAssemblyKitEndpoint")
	cmd.MarkFlagRequired("aakUsername")
	cmd.MarkFlagRequired("aakPassword")
}

// retrieve step metadata
func abapEnvironmentAssemblePackagesMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "abapEnvironmentAssemblePackages",
			Aliases:     []config.Alias{},
			Description: "Assembly of installation, support package or patch in SAP Cloud Platform ABAP Environment system",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapCredentialsId", Description: "Jenkins credentials ID containing user and password to authenticate to the Cloud Platform ABAP Environment system or the Cloud Foundry API", Type: "jenkins", Aliases: []config.Alias{{Name: "cfCredentialsId", Deprecated: false}, {Name: "credentialsId", Deprecated: false}}},
					{Name: "abapAddonAssemblyKitCredentialsId", Description: "Credential stored in Jenkins for the Addon Assembly Kit as a Service (AAKaaS) system", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "cfApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
						Default:     os.Getenv("PIPER_cfApiEndpoint"),
					},
					{
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/org"}},
						Default:     os.Getenv("PIPER_cfOrg"),
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/space"}},
						Default:     os.Getenv("PIPER_cfSpace"),
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
						Name:        "cfServiceKeyName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceKey"}, {Name: "cloudFoundry/serviceKeyName"}, {Name: "cfServiceKey"}},
						Default:     os.Getenv("PIPER_cfServiceKeyName"),
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_host"),
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapCredentialsId",
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
								Name:  "abapCredentialsId",
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
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     360,
					},
					{
						Name:        "pollIntervalsInMilliseconds",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     60000,
					},
					{
						Name:        "performAssemblePackages",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "performRegisterPackages",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
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
						Name: "aakUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapAddonAssemblyKitCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_aakUsername"),
					},
					{
						Name: "aakPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapAddonAssemblyKitCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_aakPassword"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "cf", Image: "ppiper/cf-cli:7"},
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
