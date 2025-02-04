// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
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
	ManifestVariablesFiles []string `json:"manifestVariablesFiles,omitempty"`
	CfAsync                bool     `json:"cfAsync,omitempty"`
}

// CloudFoundryCreateServiceCommand Creates one or multiple Services in Cloud Foundry
func CloudFoundryCreateServiceCommand() *cobra.Command {
	const STEP_NAME = "cloudFoundryCreateService"

	metadata := cloudFoundryCreateServiceMetadata()
	var stepConfig cloudFoundryCreateServiceOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createCloudFoundryCreateServiceCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Creates one or multiple Services in Cloud Foundry",
		Long: `Creates one or multiple Cloud Foundry Services in Cloud Foundry
Mandatory:
* Cloud Foundry API endpoint, Organization, Space and user are available

Please provide either of the following options:
* If you chose to create a single Service the Service Instance Name, Service Plan and Service Broker of the Service to be created have to be available. You can set the optional ` + "`" + `cfCreateServiceConfig` + "`" + ` flag to configure the Service creation with your respective JSON configuration. The JSON configuration can either be an in-line JSON string or the path a dedicated JSON configuration file containing the JSON configuration. If you chose a dedicated config file, you must store the file in the same folder as your ` + "`" + `Jenkinsfile` + "`" + ` that starts the Pipeline in order for the Pipeline to be able to find the file. Most favourable SCM is Git. If you want the service to be created from a particular broker you can set the optional ` + "`" + `cfServiceBroker` + "`" + `flag. You can set user provided tags for the Service creation using a flat list as the value for the optional ` + "`" + `cfServiceTags` + "`" + ` flag. The optional ` + "`" + `cfServiceBroker` + "`" + ` flag can be used when the service name is ambiguous.
* For creating one or multiple Cloud Foundry Services at once with the Cloud Foundry Create-Service-Push Plugin using the optional ` + "`" + `serviceManifest` + "`" + ` flag. If you chose to set this flag, the Create-Service-Push Plugin will be used for all Service creations in this step and you will need to provide a ` + "`" + `serviceManifest.yml` + "`" + ` file. In that case, above described flags and options will not be used for the Service creations, since you chose to use the Create-Service-Push Plugin. Please see below examples for more information on how to make use of the plugin with the appropriate step configuation. Additionally the Plugin provides the option to make use of variable substitution for the Service creations. You can find further information regarding the functionality of the Cloud Foundry Create-Service-Push Plugin in the respective documentation: [Cloud Foundry Create-Service-Push Plugin](https://github.com/dawu415/CF-CLI-Create-Service-Push-Plugin)`,
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
			cloudFoundryCreateService(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addCloudFoundryCreateServiceFlags(createCloudFoundryCreateServiceCmd, &stepConfig)
	return createCloudFoundryCreateServiceCmd
}

func addCloudFoundryCreateServiceFlags(cmd *cobra.Command, stepConfig *cloudFoundryCreateServiceOptions) {
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", `https://api.cf.eu10.hana.ondemand.com`, "Cloud Foundry API endpoint")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User or E-Mail for CF")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for Cloud Foundry User")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry org")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "Cloud Foundry Space")
	cmd.Flags().StringVar(&stepConfig.CfService, "cfService", os.Getenv("PIPER_cfService"), "Parameter for Cloud Foundry Service to be used for creating Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServicePlan, "cfServicePlan", os.Getenv("PIPER_cfServicePlan"), "Parameter for Cloud Foundry Service Plan to be used when creating a Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServiceInstanceName, "cfServiceInstanceName", os.Getenv("PIPER_cfServiceInstanceName"), "Parameter for naming the Service Instance when creating a Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfServiceBroker, "cfServiceBroker", os.Getenv("PIPER_cfServiceBroker"), "Parameter for Service Broker to be used when creating a Cloud Foundry Service")
	cmd.Flags().StringVar(&stepConfig.CfCreateServiceConfig, "cfCreateServiceConfig", os.Getenv("PIPER_cfCreateServiceConfig"), "Path to JSON file or JSON in-line string for a Cloud Foundry Service creation")
	cmd.Flags().StringVar(&stepConfig.CfServiceTags, "cfServiceTags", os.Getenv("PIPER_cfServiceTags"), "Flat list of Tags to be used when creating a Cloud Foundry Service in a single string")
	cmd.Flags().StringVar(&stepConfig.ServiceManifest, "serviceManifest", `service-manifest.yml`, "Path to Cloud Foundry Service Manifest in YAML format for multiple service creations that are being passed to a Create-Service-Push Cloud Foundry cli plugin")
	cmd.Flags().StringSliceVar(&stepConfig.ManifestVariables, "manifestVariables", []string{}, "Defines a List of variables as key-value Map objects used for variable substitution within the file given by the Manifest. Defaults to an empty list, if not specified otherwise. This can be used to set variables like it is provided by `cf push --var key=value`. The order of the maps of variables given in the list is relevant in case there are conflicting variable names and values between maps contained within the list. In case of conflicts, the last specified map in the list will win. Though each map entry in the list can contain more than one key-value pair for variable substitution, it is recommended to stick to one entry per map, and rather declare more maps within the list. The reason is that if a map in the list contains more than one key-value entry, and the entries are conflicting, the conflict resolution behavior is undefined (since map entries have no sequence). Variables defined via `manifestVariables` always win over conflicting variables defined via any file given by `manifestVariablesFiles` - no matter what is declared before. This is the same behavior as can be observed when using `cf push --var` in combination with `cf push --vars-file`")
	cmd.Flags().StringSliceVar(&stepConfig.ManifestVariablesFiles, "manifestVariablesFiles", []string{}, "Defines the manifest variables Yaml files to be used to replace variable references in manifest. This parameter is optional and will default to `manifest-variables.yml`. This can be used to set variable files like it is provided by `cf push --vars-file <file>`. If the manifest is present and so are all variable files, a variable substitution will be triggered that uses the `cfManifestSubstituteVariables` step before deployment. The format of variable references follows the Cloud Foundry standard in `https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#variable-substitution`")
	cmd.Flags().BoolVar(&stepConfig.CfAsync, "cfAsync", true, "Decides if the service creation runs asynchronously")

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
			Name:        "cloudFoundryCreateService",
			Aliases:     []config.Alias{},
			Description: "Creates one or multiple Services in Cloud Foundry",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "cfCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing user and password to authenticate to the Cloud Foundry API.", Type: "jenkins", Aliases: []config.Alias{{Name: "cloudFoundry/credentialsId", Deprecated: false}}},
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
						Name:        "cfServiceInstanceName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceInstanceName"}},
						Default:     os.Getenv("PIPER_cfServiceInstanceName"),
					},
					{
						Name:        "cfServiceBroker",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceBroker"}},
						Default:     os.Getenv("PIPER_cfServiceBroker"),
					},
					{
						Name:        "cfCreateServiceConfig",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/createServiceConfig"}},
						Default:     os.Getenv("PIPER_cfCreateServiceConfig"),
					},
					{
						Name:        "cfServiceTags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceTags"}},
						Default:     os.Getenv("PIPER_cfServiceTags"),
					},
					{
						Name:        "serviceManifest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceManifest"}, {Name: "cfServiceManifest"}},
						Default:     `service-manifest.yml`,
					},
					{
						Name:        "manifestVariables",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/manifestVariables"}, {Name: "cfManifestVariables"}},
						Default:     []string{},
					},
					{
						Name:        "manifestVariablesFiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/manifestVariablesFiles"}, {Name: "cfManifestVariablesFiles"}},
						Default:     []string{},
					},
					{
						Name:        "cfAsync",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
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
