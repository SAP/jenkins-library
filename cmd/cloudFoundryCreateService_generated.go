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

type cloudFoundryCreateServiceOptions struct {
	CfAPIEndpoint          string   `json:"cfApiEndpoint,omitempty"`
	Username               string   `json:"username,omitempty"`
	Password               string   `json:"password,omitempty"`
	CfOrg                  string   `json:"cfOrg,omitempty"`
	CfSpace                string   `json:"cfSpace,omitempty"`
	CfService              string   `json:"cfService,omitempty"`
	CfServicePlan          string   `json:"cfServicePlan,omitempty"`
	CfServiceInstanceName  string   `json:"cfServiceInstanceName,omitempty"`
	CfServiceBroker        string   `json:"cfServiceBroker,omitempty"`
	CfCreateServiceConfig  string   `json:"cfCreateServiceConfig,omitempty"`
	CfServiceTags          string   `json:"cfServiceTags,omitempty"`
	ServiceManifest        string   `json:"serviceManifest,omitempty"`
	ManifestVariables      []string `json:"manifestVariables,omitempty"`
	ManifestVariablesFiles string   `json:"manifestVariablesFiles,omitempty"`
}

// CloudFoundryCreateServiceCommand cloudFoundryCreateService
func CloudFoundryCreateServiceCommand() *cobra.Command {
	const STEP_NAME = "cloudFoundryCreateService"

	metadata := cloudFoundryCreateServiceMetadata()
	var stepConfig cloudFoundryCreateServiceOptions
	var startTime time.Time

	var createCloudFoundryCreateServiceCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "cloudFoundryCreateService",
		Long:  `Create CloudFoundryService`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
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
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			cloudFoundryCreateService(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addCloudFoundryCreateServiceFlags(createCloudFoundryCreateServiceCmd, &stepConfig)
	return createCloudFoundryCreateServiceCmd
}

func addCloudFoundryCreateServiceFlags(cmd *cobra.Command, stepConfig *cloudFoundryCreateServiceOptions) {
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", os.Getenv("PIPER_cfApiEndpoint"), "Cloud Foundry API endpoint")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User or E-Mail for CF")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "User Password for CF User")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "CF org")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "CF Space")
	cmd.Flags().StringVar(&stepConfig.CfService, "cfService", os.Getenv("PIPER_cfService"), "Parameter for Cloud Foundry Service to be used for creating Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServicePlan, "cfServicePlan", os.Getenv("PIPER_cfServicePlan"), "Parameter for Cloud Foundry Service Plan to be used when creating Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServiceInstanceName, "cfServiceInstanceName", os.Getenv("PIPER_cfServiceInstanceName"), "Parameter for naming the Service Instance when creating Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServiceBroker, "cfServiceBroker", os.Getenv("PIPER_cfServiceBroker"), "Parameter for Service Broker to be used when creating Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfCreateServiceConfig, "cfCreateServiceConfig", os.Getenv("PIPER_cfCreateServiceConfig"), "Path to JSON file or JSON in-line String for Cloud Foundry Service creation")
	cmd.Flags().StringVar(&stepConfig.CfServiceTags, "cfServiceTags", os.Getenv("PIPER_cfServiceTags"), "Flat list of Tags to be used when creating Cloud Foundry Service in a single string")
	cmd.Flags().StringVar(&stepConfig.ServiceManifest, "serviceManifest", os.Getenv("PIPER_serviceManifest"), "Path to Cloud Foundry service manifest in YAML format for multiple service creations that are being passed to a Create-Service-Push cf cli plugin")
	cmd.Flags().StringSliceVar(&stepConfig.ManifestVariables, "manifestVariables", []string{}, "Defines a List of variables as key-value Map objects used for variable substitution within the file given by manifest. Defaults to an empty list, if not specified otherwise. This can be used to set variables like it is provided by `cf push --var key=value`. The order of the maps of variables given in the list is relevant in case there are conflicting variable names and values between maps contained within the list. In case of conflicts, the last specified map in the list will win. Though each map entry in the list can contain more than one key-value pair for variable substitution, it is recommended to stick to one entry per map, and rather declare more maps within the list. The reason is that if a map in the list contains more than one key-value entry, and the entries are conflicting, the conflict resolution behavior is undefined (since map entries have no sequence). Variables defined via `manifestVariables` always win over conflicting variables defined via any file given by `manifestVariablesFiles` - no matter what is declared before. This is the same behavior as can be observed when using `cf push --var` in combination with `cf push --vars-file`")
	cmd.Flags().StringVar(&stepConfig.ManifestVariablesFiles, "manifestVariablesFiles", os.Getenv("PIPER_manifestVariablesFiles"), "Defines the manifest variables Yaml files to be used to replace variable references in manifest. This parameter is optional and will default to `manifest-variables.yml`. This can be used to set variable files like it is provided by `cf push --vars-file <file>`. If the manifest is present and so are all variable files, a variable substitution will be triggered that uses the `cfManifestSubstituteVariables` step before deployment. The format of variable references follows the Cloud Foundry standard in `https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#variable-substitution`")

	cmd.MarkFlagRequired("cfApiEndpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("cfOrg")
	cmd.MarkFlagRequired("cfSpace")
}

// retrieve step metadata
func cloudFoundryCreateServiceMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "cloudFoundryCreateService",
			Aliases: []config.Alias{},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "cfApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
					},
					{
						Name:        "username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/org"}},
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cloudFoundry/space"}},
					},
					{
						Name:        "cfService",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/service"}},
					},
					{
						Name:        "cfServicePlan",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/servicePlan"}},
					},
					{
						Name:        "cfServiceInstanceName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceInstanceName"}},
					},
					{
						Name:        "cfServiceBroker",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceBroker"}},
					},
					{
						Name:        "cfCreateServiceConfig",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/createServiceConfig"}},
					},
					{
						Name:        "cfServiceTags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceTags"}},
					},
					{
						Name:        "serviceManifest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceManifest"}},
					},
					{
						Name:        "manifestVariables",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/manifestVariables"}},
					},
					{
						Name:        "manifestVariablesFiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/manifestVariablesFiles"}},
					},
				},
			},
		},
	}
	return theMetaData
}
