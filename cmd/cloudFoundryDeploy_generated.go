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

type cloudFoundryDeployOptions struct {
	APIEndpoint              string   `json:"apiEndpoint,omitempty"`
	AppName                  string   `json:"appName,omitempty"`
	ArtifactVersion          string   `json:"artifactVersion,omitempty"`
	CfHome                   string   `json:"cfHome,omitempty"`
	CfNativeDeployParameters string   `json:"cfNativeDeployParameters,omitempty"`
	CfPluginHome             string   `json:"cfPluginHome,omitempty"`
	DeployDockerImage        string   `json:"deployDockerImage,omitempty"`
	DeployTool               string   `json:"deployTool,omitempty"`
	BuildTool                string   `json:"buildTool,omitempty"`
	DeployType               string   `json:"deployType,omitempty"`
	DockerPassword           string   `json:"dockerPassword,omitempty"`
	DockerUsername           string   `json:"dockerUsername,omitempty"`
	KeepOldInstance          bool     `json:"keepOldInstance,omitempty"`
	LoginParameters          string   `json:"loginParameters,omitempty"`
	Manifest                 string   `json:"manifest,omitempty"`
	ManifestVariables        []string `json:"manifestVariables,omitempty"`
	ManifestVariablesFiles   []string `json:"manifestVariablesFiles,omitempty"`
	MtaDeployParameters      string   `json:"mtaDeployParameters,omitempty"`
	MtaExtensionDescriptor   string   `json:"mtaExtensionDescriptor,omitempty"`
	MtaPath                  string   `json:"mtaPath,omitempty"`
	Org                      string   `json:"org,omitempty"`
	Password                 string   `json:"password,omitempty"`
	SmokeTestScript          string   `json:"smokeTestScript,omitempty"`
	SmokeTestStatusCode      int      `json:"smokeTestStatusCode,omitempty"`
	Space                    string   `json:"space,omitempty"`
	Username                 string   `json:"username,omitempty"`
}

type cloudFoundryDeployInflux struct {
	deployment_data struct {
		fields struct {
			artifactURL string
			deployTime  string
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
		value       string
	}{
		{valType: config.InfluxField, measurement: "deployment_data", name: "artifactUrl", value: i.deployment_data.fields.artifactURL},
		{valType: config.InfluxField, measurement: "deployment_data", name: "deployTime", value: i.deployment_data.fields.deployTime},
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
		log.Entry().Fatal("failed to persist Influx environment")
	}
}

// CloudFoundryDeployCommand Deploys an application to Cloud Foundry
func CloudFoundryDeployCommand() *cobra.Command {
	const STEP_NAME = "cloudFoundryDeploy"

	metadata := cloudFoundryDeployMetadata()
	var stepConfig cloudFoundryDeployOptions
	var startTime time.Time
	var influx cloudFoundryDeployInflux

	var createCloudFoundryDeployCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Deploys an application to Cloud Foundry",
		Long:  `Deploys an application to a test or production space within Cloud Foundry.`,
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
			log.RegisterSecret(stepConfig.DockerPassword)
			log.RegisterSecret(stepConfig.Password)
			log.RegisterSecret(stepConfig.Username)

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
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			cloudFoundryDeploy(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
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
	cmd.Flags().StringVar(&stepConfig.CfHome, "cfHome", os.Getenv("PIPER_cfHome"), "The cf home folder used by the cf cli. If not provided the default assumed by the cf cli is used.")
	cmd.Flags().StringVar(&stepConfig.CfNativeDeployParameters, "cfNativeDeployParameters", os.Getenv("PIPER_cfNativeDeployParameters"), "Additional parameters passed to cf native deployment command")
	cmd.Flags().StringVar(&stepConfig.CfPluginHome, "cfPluginHome", os.Getenv("PIPER_cfPluginHome"), "The cf plugin home folder used by the cf cli. If not provided the default assumed by the cf cli is used.")
	cmd.Flags().StringVar(&stepConfig.DeployDockerImage, "deployDockerImage", os.Getenv("PIPER_deployDockerImage"), "Docker image deployments are supported (via manifest file in general)[https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#docker]. If no manifest is used, this parameter defines the image to be deployed. The specified name of the image is passed to the `--docker-image` parameter of the cf CLI and must adhere it's naming pattern (e.g. REPO/IMAGE:TAG). See (cf CLI documentation)[https://docs.cloudfoundry.org/devguide/deploy-apps/push-docker.html] for details. Note: The used Docker registry must be visible for the targeted Cloud Foundry instance.")
	cmd.Flags().StringVar(&stepConfig.DeployTool, "deployTool", os.Getenv("PIPER_deployTool"), "Defines the tool which should be used for deployment.")
	cmd.Flags().StringVar(&stepConfig.BuildTool, "buildTool", os.Getenv("PIPER_buildTool"), "Defines the tool which is used for building the artifact. If provided, `deployTool` is automatically derived from it. For MTA projects, `deployTool` defaults to `mtaDeployPlugin`. For other projects `cf_native` will be used.")
	cmd.Flags().StringVar(&stepConfig.DeployType, "deployType", `standard`, "Defines the type of deployment, either `standard` deployment which results in a system downtime or a zero-downtime `blue-green` deployment.If 'cf_native' as deployType and 'blue-green' as deployTool is used in combination, your manifest.yaml may only contain one application. If this application has the option 'no-route' active the deployType will be changed to 'standard'.")
	cmd.Flags().StringVar(&stepConfig.DockerPassword, "dockerPassword", os.Getenv("PIPER_dockerPassword"), "If the specified image in `deployDockerImage` is contained in a Docker registry, which requires authorization, this defines the password to be used.")
	cmd.Flags().StringVar(&stepConfig.DockerUsername, "dockerUsername", os.Getenv("PIPER_dockerUsername"), "If the specified image in `deployDockerImage` is contained in a Docker registry, which requires authorization, this defines the username to be used.")
	cmd.Flags().BoolVar(&stepConfig.KeepOldInstance, "keepOldInstance", false, "In case of a `blue-green` deployment the old instance will be deleted by default. If this option is set to true the old instance will remain stopped in the Cloud Foundry space.")
	cmd.Flags().StringVar(&stepConfig.LoginParameters, "loginParameters", os.Getenv("PIPER_loginParameters"), "Addition command line options for cf login command. No escaping/quoting is performed. Not recommended for productive environments.")
	cmd.Flags().StringVar(&stepConfig.Manifest, "manifest", os.Getenv("PIPER_manifest"), "Defines the manifest to be used for deployment to Cloud Foundry.")
	cmd.Flags().StringSliceVar(&stepConfig.ManifestVariables, "manifestVariables", []string{}, "Defines a list of variables as key-value Map objects used for variable substitution within the file given by manifest. Defaults to an empty list, if not specified otherwise. This can be used to set variables like it is provided by 'cf push --var key=value'. The order of the maps of variables given in the list is relevant in case there are conflicting variable names and value between maps contained within the list. In case of conflicts, the last specified map in the list will win. Though each map entry in the list can contain more than one key-value pair for variable substitution, it is recommended to stick to one entry per map, and rather declare more maps within the list. The reason is that if a map in the list contains more than one key-value entry, and the entries are conflicting, the conflict resolution behavior is undefined (since map entries have no sequence). Note: variables defined via 'manifestVariables' always win over conflicting variables defined via any file given by 'manifestVariablesFiles' - no matter what is declared before. This is the same behavior as can be observed when using 'cf push --var' in combination with 'cf push --vars-file'.")
	cmd.Flags().StringSliceVar(&stepConfig.ManifestVariablesFiles, "manifestVariablesFiles", []string{`manifest-variables.yml`}, "path(s) of the Yaml file(s) containing the variable values to use as a replacement in the manifest file. The order of the files is relevant in case there are conflicting variable names and values within variable files. In such a case, the values of the last file win.")
	cmd.Flags().StringVar(&stepConfig.MtaDeployParameters, "mtaDeployParameters", `-f`, "Additional parameters passed to mta deployment command")
	cmd.Flags().StringVar(&stepConfig.MtaExtensionDescriptor, "mtaExtensionDescriptor", os.Getenv("PIPER_mtaExtensionDescriptor"), "Defines additional extension descriptor file for deployment with the mtaDeployPlugin")
	cmd.Flags().StringVar(&stepConfig.MtaPath, "mtaPath", os.Getenv("PIPER_mtaPath"), "Defines the path to *.mtar for deployment with the mtaDeployPlugin")
	cmd.Flags().StringVar(&stepConfig.Org, "org", os.Getenv("PIPER_org"), "Cloud Foundry target organization.")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password")
	cmd.Flags().StringVar(&stepConfig.SmokeTestScript, "smokeTestScript", `blueGreenCheckScript.sh`, "Allows to specify a script which performs a check during blue-green deployment. The script gets the FQDN as parameter and returns `exit code 0` in case check returned `smokeTestStatusCode`. More details can be found [here](https://github.com/bluemixgaragelondon/cf-blue-green-deploy#how-to-use). Currently this option is only considered for deployTool `cf_native`.")
	cmd.Flags().IntVar(&stepConfig.SmokeTestStatusCode, "smokeTestStatusCode", 200, "Expected status code returned by the check.")
	cmd.Flags().StringVar(&stepConfig.Space, "space", os.Getenv("PIPER_space"), "Cloud Foundry target space")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User")

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
			Name:    "cloudFoundryDeploy",
			Aliases: []config.Alias{},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "apiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cfApiEndpoint"}, {Name: "cloudFoundry/apiEndpoint"}},
					},
					{
						Name:        "appName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfAppName"}, {Name: "cloudFoundry/appName"}},
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
					},
					{
						Name:        "cfHome",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfNativeDeployParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cfPluginHome",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "deployDockerImage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "deployTool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
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
					},
					{
						Name:        "deployType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
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
					},
					{
						Name:        "keepOldInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "loginParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "manifest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfManifest"}, {Name: "cloudFoundry/manifest"}},
					},
					{
						Name:        "manifestVariables",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfManifestVariables"}, {Name: "cloudFoundry/manifestVariables"}},
					},
					{
						Name:        "manifestVariablesFiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cfManifestVariablesFiles"}, {Name: "cloudFoundry/manifestVariablesFiles"}},
					},
					{
						Name:        "mtaDeployParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "mtaExtensionDescriptor",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/mtaExtensionDescriptor"}},
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
					},
					{
						Name:        "org",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cfOrg"}, {Name: "cloudFoundry/org"}},
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cfCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:        "smokeTestScript",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "smokeTestStatusCode",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "space",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "cfSpace"}, {Name: "cloudFoundry/space"}},
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cfCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
