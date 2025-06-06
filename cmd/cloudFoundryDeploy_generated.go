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

type cloudFoundryDeployOptions struct {
	APIEndpoint              string                 `json:"apiEndpoint,omitempty"`
	AppName                  string                 `json:"appName,omitempty"`
	ArtifactVersion          string                 `json:"artifactVersion,omitempty"`
	CommitHash               string                 `json:"commitHash,omitempty"`
	CfHome                   string                 `json:"cfHome,omitempty"`
	CfNativeDeployParameters string                 `json:"cfNativeDeployParameters,omitempty"`
	CfPluginHome             string                 `json:"cfPluginHome,omitempty"`
	DeployDockerImage        string                 `json:"deployDockerImage,omitempty"`
	DeployTool               string                 `json:"deployTool,omitempty"`
	BuildTool                string                 `json:"buildTool,omitempty"`
	DeployType               string                 `json:"deployType,omitempty"`
	DockerPassword           string                 `json:"dockerPassword,omitempty"`
	DockerUsername           string                 `json:"dockerUsername,omitempty"`
	KeepOldInstance          bool                   `json:"keepOldInstance,omitempty"`
	LoginParameters          string                 `json:"loginParameters,omitempty"`
	Manifest                 string                 `json:"manifest,omitempty"`
	ManifestVariables        []string               `json:"manifestVariables,omitempty"`
	ManifestVariablesFiles   []string               `json:"manifestVariablesFiles,omitempty"`
	MtaDeployParameters      string                 `json:"mtaDeployParameters,omitempty"`
	MtaExtensionDescriptor   string                 `json:"mtaExtensionDescriptor,omitempty"`
	MtaExtensionCredentials  map[string]interface{} `json:"mtaExtensionCredentials,omitempty"`
	MtaPath                  string                 `json:"mtaPath,omitempty"`
	Org                      string                 `json:"org,omitempty"`
	Password                 string                 `json:"password,omitempty"`
	Space                    string                 `json:"space,omitempty"`
	Username                 string                 `json:"username,omitempty"`
}

type cloudFoundryDeployInflux struct {
	deployment_data struct {
		fields struct {
			artifactURL string
			deployTime  string
			commitHash  string
			jobTrigger  string
		}
		tags struct {
			artifactVersion string
			deployUser      string
			deployResult    string
			cfAPIEndpoint   string
			cfOrg           string
			cfSpace         string
		}
	}
}

func (i *cloudFoundryDeployInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "deployment_data", name: "artifactUrl", value: i.deployment_data.fields.artifactURL},
		{valType: config.InfluxField, measurement: "deployment_data", name: "deployTime", value: i.deployment_data.fields.deployTime},
		{valType: config.InfluxField, measurement: "deployment_data", name: "commitHash", value: i.deployment_data.fields.commitHash},
		{valType: config.InfluxField, measurement: "deployment_data", name: "jobTrigger", value: i.deployment_data.fields.jobTrigger},
		{valType: config.InfluxTag, measurement: "deployment_data", name: "artifactVersion", value: i.deployment_data.tags.artifactVersion},
		{valType: config.InfluxTag, measurement: "deployment_data", name: "deployUser", value: i.deployment_data.tags.deployUser},
		{valType: config.InfluxTag, measurement: "deployment_data", name: "deployResult", value: i.deployment_data.tags.deployResult},
		{valType: config.InfluxTag, measurement: "deployment_data", name: "cfApiEndpoint", value: i.deployment_data.tags.cfAPIEndpoint},
		{valType: config.InfluxTag, measurement: "deployment_data", name: "cfOrg", value: i.deployment_data.tags.cfOrg},
		{valType: config.InfluxTag, measurement: "deployment_data", name: "cfSpace", value: i.deployment_data.tags.cfSpace},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Influx environment")
	}
}

// CloudFoundryDeployCommand Deploys an application to Cloud Foundry
func CloudFoundryDeployCommand() *cobra.Command {
	const STEP_NAME = "cloudFoundryDeploy"

	metadata := cloudFoundryDeployMetadata()
	var stepConfig cloudFoundryDeployOptions
	var startTime time.Time
	var influx cloudFoundryDeployInflux
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createCloudFoundryDeployCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Deploys an application to Cloud Foundry",
		Long: `Deploys an application to a test or production space within Cloud Foundry.
This step supports two deployment types:

* in a standard way
* in a zero-downtime manner using a [blue-green deployment approach](https://martinfowler.com/bliki/BlueGreenDeployment.html)

The step achieves this via following deploy tools
* [cf CLI](https://docs.cloudfoundry.org/cf-cli/) - used as default for Non MTA apps
* [MTA CF CLI Plugin](https://github.com/cloudfoundry-incubator/multiapps-cli-plugin) - used as default for MTA apps`,
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
			log.RegisterSecret(stepConfig.DockerPassword)
			log.RegisterSecret(stepConfig.DockerUsername)
			log.RegisterSecret(stepConfig.Password)
			log.RegisterSecret(stepConfig.Username)

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
				influx.persist(GeneralConfig.EnvRootPath, "influx")
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
			cloudFoundryDeploy(stepConfig, &stepTelemetryData, &influx)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addCloudFoundryDeployFlags(createCloudFoundryDeployCmd, &stepConfig)
	return createCloudFoundryDeployCmd
}

func addCloudFoundryDeployFlags(cmd *cobra.Command, stepConfig *cloudFoundryDeployOptions) {
	cmd.Flags().StringVar(&stepConfig.APIEndpoint, "apiEndpoint", `https://api.cf.eu10.hana.ondemand.com`, "Cloud Foundry API endpoint")
	cmd.Flags().StringVar(&stepConfig.AppName, "appName", os.Getenv("PIPER_appName"), "Defines the name of the application to be deployed to the Cloud Foundry space")
	cmd.Flags().StringVar(&stepConfig.ArtifactVersion, "artifactVersion", os.Getenv("PIPER_artifactVersion"), "The artifact version, used for influx reporting")
	cmd.Flags().StringVar(&stepConfig.CommitHash, "commitHash", os.Getenv("PIPER_commitHash"), "The commit hash, used for influx reporting")
	cmd.Flags().StringVar(&stepConfig.CfHome, "cfHome", os.Getenv("PIPER_cfHome"), "The cf home folder used by the cf cli. If not provided the default assumed by the cf cli is used.")
	cmd.Flags().StringVar(&stepConfig.CfNativeDeployParameters, "cfNativeDeployParameters", os.Getenv("PIPER_cfNativeDeployParameters"), "Additional parameters passed to cf native deployment command")
	cmd.Flags().StringVar(&stepConfig.CfPluginHome, "cfPluginHome", os.Getenv("PIPER_cfPluginHome"), "The cf plugin home folder used by the cf cli. If not provided the default assumed by the cf cli is used.")
	cmd.Flags().StringVar(&stepConfig.DeployDockerImage, "deployDockerImage", os.Getenv("PIPER_deployDockerImage"), "Docker image deployments are supported [via manifest file in general](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#docker). If no manifest is used, this parameter defines the image to be deployed. The specified name of the image is passed to the `--docker-image` parameter of the cf CLI and must adhere it's naming pattern (e.g. REPO/IMAGE:TAG). See [cf CLI documentation](https://docs.cloudfoundry.org/devguide/deploy-apps/push-docker.html)x`x` for details. Note: The used Docker registry must be visible for the targeted Cloud Foundry instance.")
	cmd.Flags().StringVar(&stepConfig.DeployTool, "deployTool", os.Getenv("PIPER_deployTool"), "Defines the tool which should be used for deployment. Mandatory if `buildTool` is not found in pipeline environment")
	cmd.Flags().StringVar(&stepConfig.BuildTool, "buildTool", os.Getenv("PIPER_buildTool"), "Defines the tool which is used for building the artifact. If provided, `deployTool` is automatically derived from it. For MTA projects, `deployTool` defaults to `mtaDeployPlugin`. For other projects `cf_native` will be used.")
	cmd.Flags().StringVar(&stepConfig.DeployType, "deployType", `standard`, "Defines the type of deployment -`standard` or `blue-green` deployment. For mta build tool, possible values are `standard`, `blue-green` or `bg-deploy`. For cf native build tools, possible value is `standard`. To eliminate system downtime, an alternative is to pass '--strategy rolling' to the parameter `cfNativeDeployParameters`.")
	cmd.Flags().StringVar(&stepConfig.DockerPassword, "dockerPassword", os.Getenv("PIPER_dockerPassword"), "If the specified image in `deployDockerImage` is contained in a Docker registry, which requires authorization, this defines the password to be used.")
	cmd.Flags().StringVar(&stepConfig.DockerUsername, "dockerUsername", os.Getenv("PIPER_dockerUsername"), "If the specified image in `deployDockerImage` is contained in a Docker registry, which requires authorization, this defines the username to be used.")
	cmd.Flags().BoolVar(&stepConfig.KeepOldInstance, "keepOldInstance", false, "If this option is set to true the old instance will remain stopped in the Cloud Foundry space.\"")
	cmd.Flags().StringVar(&stepConfig.LoginParameters, "loginParameters", os.Getenv("PIPER_loginParameters"), "Addition command line options for cf login command. No escaping/quoting is performed. Not recommended for productive environments.")
	cmd.Flags().StringVar(&stepConfig.Manifest, "manifest", os.Getenv("PIPER_manifest"), "Defines the manifest file name to be used for deployment to Cloud Foundry. Defaults to `manifest.yml`")
	cmd.Flags().StringSliceVar(&stepConfig.ManifestVariables, "manifestVariables", []string{}, "Defines a list of variables in the form `key=value` which are used for variable substitution within the file given by manifest.")
	cmd.Flags().StringSliceVar(&stepConfig.ManifestVariablesFiles, "manifestVariablesFiles", []string{`manifest-variables.yml`}, "path(s) of the Yaml file(s) containing the variable values to use as a replacement in the manifest file. The order of the files is relevant in case there are conflicting variable names and values within variable files. In such a case, the values of the last file win.")
	cmd.Flags().StringVar(&stepConfig.MtaDeployParameters, "mtaDeployParameters", `-f`, "Additional parameters passed to mta deployment command")
	cmd.Flags().StringVar(&stepConfig.MtaExtensionDescriptor, "mtaExtensionDescriptor", os.Getenv("PIPER_mtaExtensionDescriptor"), "Defines additional extension descriptor file for deployment with the mtaDeployPlugin")

	cmd.Flags().StringVar(&stepConfig.MtaPath, "mtaPath", os.Getenv("PIPER_mtaPath"), "Defines the path to *.mtar for deployment with the mtaDeployPlugin")
	cmd.Flags().StringVar(&stepConfig.Org, "org", os.Getenv("PIPER_org"), "Cloud Foundry target organization.")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password")
	cmd.Flags().StringVar(&stepConfig.Space, "space", os.Getenv("PIPER_space"), "Cloud Foundry target space")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User name used for deployment")

	cmd.MarkFlagRequired("apiEndpoint")
	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("username")
}

// retrieve step metadata
func cloudFoundryDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "cloudFoundryDeploy",
			Aliases:     []config.Alias{},
			Description: "Deploys an application to Cloud Foundry",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "cfCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing user and password to authenticate to the Cloud Foundry API.", Type: "jenkins", Aliases: []config.Alias{{Name: "cloudFoundry/credentialsId", Deprecated: true}}},
					{Name: "dockerCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing user and password to authenticate to the Docker registry.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "apiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cfApiEndpoint"}, {Name: "cloudFoundry/apiEndpoint", Deprecated: true}},
						Default:     `https://api.cf.eu10.hana.ondemand.com`,
					},
					{
						Name:        "appName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfAppName"}, {Name: "cloudFoundry/appName", Deprecated: true}},
						Default:     os.Getenv("PIPER_appName"),
					},
					{
						Name: "artifactVersion",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_artifactVersion"),
					},
					{
						Name: "commitHash",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "git/headCommitId",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_commitHash"),
					},
					{
						Name:        "cfHome",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_cfHome"),
					},
					{
						Name:        "cfNativeDeployParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_cfNativeDeployParameters"),
					},
					{
						Name:        "cfPluginHome",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_cfPluginHome"),
					},
					{
						Name:        "deployDockerImage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_deployDockerImage"),
					},
					{
						Name:        "deployTool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_deployTool"),
					},
					{
						Name: "buildTool",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "buildTool",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_buildTool"),
					},
					{
						Name:        "deployType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `standard`,
					},
					{
						Name: "dockerPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "dockerCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_dockerPassword"),
					},
					{
						Name: "dockerUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "dockerCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_dockerUsername"),
					},
					{
						Name:        "keepOldInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "loginParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_loginParameters"),
					},
					{
						Name:        "manifest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfManifest"}, {Name: "cloudFoundry/manifest", Deprecated: true}},
						Default:     os.Getenv("PIPER_manifest"),
					},
					{
						Name:        "manifestVariables",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfManifestVariables"}, {Name: "cloudFoundry/manifestVariables", Deprecated: true}},
						Default:     []string{},
					},
					{
						Name:        "manifestVariablesFiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfManifestVariablesFiles"}, {Name: "cloudFoundry/manifestVariablesFiles", Deprecated: true}},
						Default:     []string{`manifest-variables.yml`},
					},
					{
						Name:        "mtaDeployParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `-f`,
					},
					{
						Name:        "mtaExtensionDescriptor",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/mtaExtensionDescriptor", Deprecated: true}},
						Default:     os.Getenv("PIPER_mtaExtensionDescriptor"),
					},
					{
						Name:        "mtaExtensionCredentials",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/mtaExtensionCredentials", Deprecated: true}},
					},
					{
						Name: "mtaPath",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "mtarFilePath",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_mtaPath"),
					},
					{
						Name:        "org",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cfOrg"}, {Name: "cloudFoundry/org", Deprecated: true}},
						Default:     os.Getenv("PIPER_org"),
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
						Name:        "space",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cfSpace"}, {Name: "cloudFoundry/space", Deprecated: true}},
						Default:     os.Getenv("PIPER_space"),
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
				},
			},
			Containers: []config.Container{
				{Name: "cfDeploy", Image: "ppiper/cf-cli:latest", Options: []config.Option{{Name: "--ulimit", Value: "stack=67108864:67108864"}, {Name: "--ulimit", Value: "nofile=65536:65536"}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"name": "deployment_data", "fields": []map[string]string{{"name": "artifactUrl"}, {"name": "deployTime"}, {"name": "commitHash"}, {"name": "jobTrigger"}}, "tags": []map[string]string{{"name": "artifactVersion"}, {"name": "deployUser"}, {"name": "deployResult"}, {"name": "cfApiEndpoint"}, {"name": "cfOrg"}, {"name": "cfSpace"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
