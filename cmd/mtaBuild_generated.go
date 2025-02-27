// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"
)

type mtaBuildOptions struct {
	MtarName                        string   `json:"mtarName,omitempty"`
	MtarGroup                       string   `json:"mtarGroup,omitempty"`
	Version                         string   `json:"version,omitempty"`
	Extensions                      string   `json:"extensions,omitempty"`
	Jobs                            int      `json:"jobs,omitempty"`
	Platform                        string   `json:"platform,omitempty" validate:"possible-values=CF NEO XSA"`
	ApplicationName                 string   `json:"applicationName,omitempty"`
	Source                          string   `json:"source,omitempty"`
	Target                          string   `json:"target,omitempty"`
	DefaultNpmRegistry              string   `json:"defaultNpmRegistry,omitempty"`
	ProjectSettingsFile             string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile              string   `json:"globalSettingsFile,omitempty"`
	M2Path                          string   `json:"m2Path,omitempty"`
	InstallArtifacts                bool     `json:"installArtifacts,omitempty"`
	MtaDeploymentRepositoryPassword string   `json:"mtaDeploymentRepositoryPassword,omitempty"`
	MtaDeploymentRepositoryUser     string   `json:"mtaDeploymentRepositoryUser,omitempty"`
	MtaDeploymentRepositoryURL      string   `json:"mtaDeploymentRepositoryUrl,omitempty"`
	Publish                         bool     `json:"publish,omitempty"`
	Profiles                        []string `json:"profiles,omitempty"`
	BuildSettingsInfo               string   `json:"buildSettingsInfo,omitempty"`
	CreateBOM                       bool     `json:"createBOM,omitempty"`
	EnableSetTimestamp              bool     `json:"enableSetTimestamp,omitempty"`
	CreateBuildArtifactsMetadata    bool     `json:"createBuildArtifactsMetadata,omitempty"`
}

type mtaBuildCommonPipelineEnvironment struct {
	mtarFilePath string
	custom       struct {
		mtaBuildToolDesc  string
		mtarPublishedURL  string
		buildSettingsInfo string
		mtaBuildArtifacts string
	}
}

func (p *mtaBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "", name: "mtarFilePath", value: p.mtarFilePath},
		{category: "custom", name: "mtaBuildToolDesc", value: p.custom.mtaBuildToolDesc},
		{category: "custom", name: "mtarPublishedUrl", value: p.custom.mtarPublishedURL},
		{category: "custom", name: "buildSettingsInfo", value: p.custom.buildSettingsInfo},
		{category: "custom", name: "mtaBuildArtifacts", value: p.custom.mtaBuildArtifacts},
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

type mtaBuildReports struct {
}

func (p *mtaBuildReports) persist(stepConfig mtaBuildOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/TEST-*.xml", ParamRef: "", StepResultType: "junit"},
		{FilePattern: "**/cobertura-coverage.xml", ParamRef: "", StepResultType: "cobertura-coverage"},
		{FilePattern: "**/jacoco.xml", ParamRef: "", StepResultType: "jacoco-coverage"},
	}
	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcpJsonKeyFilePath, Modified: false},
	}
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
		return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
	}
}

// MtaBuildCommand Performs an mta build
func MtaBuildCommand() *cobra.Command {
	const STEP_NAME = "mtaBuild"

	metadata := mtaBuildMetadata()
	var stepConfig mtaBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment mtaBuildCommonPipelineEnvironment
	var reports mtaBuildReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createMtaBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Performs an mta build",
		Long: `Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.
### build with depedencies from a private repository
1. For maven related settings refer [maven build dependencies](./mavenBuild.md#build-with-depedencies-from-a-private-repository)
2. For NPM related settings refer [NPM build dependencies](./npmExecuteScripts.md#build-with-depedencies-from-a-private-repository)`,
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
			log.RegisterSecret(stepConfig.MtaDeploymentRepositoryPassword)

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
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
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
			mtaBuild(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addMtaBuildFlags(createMtaBuildCmd, &stepConfig)
	return createMtaBuildCmd
}

func addMtaBuildFlags(cmd *cobra.Command, stepConfig *mtaBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.MtarName, "mtarName", os.Getenv("PIPER_mtarName"), "The name of the generated mtar file including its extension.")
	cmd.Flags().StringVar(&stepConfig.MtarGroup, "mtarGroup", os.Getenv("PIPER_mtarGroup"), "The group to which the mtar artifact will be uploaded. Required when publish is True.")
	cmd.Flags().StringVar(&stepConfig.Version, "version", os.Getenv("PIPER_version"), "Version of the mtar artifact")
	cmd.Flags().StringVar(&stepConfig.Extensions, "extensions", os.Getenv("PIPER_extensions"), "The path to the extension descriptor file.")
	cmd.Flags().IntVar(&stepConfig.Jobs, "jobs", 0, "Configures the number of Make jobs that can run simultaneously. Maximum value allowed is 8")
	cmd.Flags().StringVar(&stepConfig.Platform, "platform", `CF`, "The target platform to which the mtar can be deployed.")
	cmd.Flags().StringVar(&stepConfig.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "The name of the application which is being built. If the parameter has been provided and no `mta.yaml` exists, the `mta.yaml` will be automatically generated using this parameter and the information (`name` and `version`) from 'package.json` before the actual build starts.")
	cmd.Flags().StringVar(&stepConfig.Source, "source", `./`, "The path to the MTA project.")
	cmd.Flags().StringVar(&stepConfig.Target, "target", `./`, "The folder for the generated `MTAR` file. If the parameter has been provided, the `MTAR` file is saved in the root of the folder provided by the argument.")
	cmd.Flags().StringVar(&stepConfig.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "Url to the npm registry that should be used for installing npm dependencies.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path or url to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path or url to the mvn settings file that should be used as global settings file")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")
	cmd.Flags().BoolVar(&stepConfig.InstallArtifacts, "installArtifacts", false, "If enabled, for npm packages this step will install all dependencies including dev dependencies. For maven it will install all artifacts to the local maven repository. Note: This happens _after_ mta build was done. The default mta build tool does not install dev-dependencies as part of the process. If you require dev-dependencies for building the mta, you will need to use a [custom builder](https://sap.github.io/cloud-mta-build-tool/configuration/#configuring-the-custom-builder)")
	cmd.Flags().StringVar(&stepConfig.MtaDeploymentRepositoryPassword, "mtaDeploymentRepositoryPassword", os.Getenv("PIPER_mtaDeploymentRepositoryPassword"), "Password for the alternative deployment repository to which mtar artifacts will be publised")
	cmd.Flags().StringVar(&stepConfig.MtaDeploymentRepositoryUser, "mtaDeploymentRepositoryUser", os.Getenv("PIPER_mtaDeploymentRepositoryUser"), "User for the alternative deployment repository to which which mtar artifacts will be publised")
	cmd.Flags().StringVar(&stepConfig.MtaDeploymentRepositoryURL, "mtaDeploymentRepositoryUrl", os.Getenv("PIPER_mtaDeploymentRepositoryUrl"), "Url for the alternative deployment repository to which mtar artifacts will be publised")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "pushed mtar artifact to altDeploymentRepositoryUrl/altDeploymentRepositoryID when set to true")
	cmd.Flags().StringSliceVar(&stepConfig.Profiles, "profiles", []string{}, "Defines list of maven build profiles to be used. profiles will overwrite existing values in the global settings xml at $M2_HOME/conf/settings.xml")
	cmd.Flags().StringVar(&stepConfig.BuildSettingsInfo, "buildSettingsInfo", os.Getenv("PIPER_buildSettingsInfo"), "build settings info is typically filled by the step automatically to create information about the build settings that were used during the mta build . This information is typically used for compliance related processes.")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Creates the bill of materials (BOM) using CycloneDX plugin.")
	cmd.Flags().BoolVar(&stepConfig.EnableSetTimestamp, "enableSetTimestamp", true, "Enables setting the timestamp in the `mta.yaml` when it contains `${timestamp}`. Disable this when you want the MTA Deploy Service to do this instead.")
	cmd.Flags().BoolVar(&stepConfig.CreateBuildArtifactsMetadata, "createBuildArtifactsMetadata", false, "metadata about the artifacts that are build and published, this metadata is generally used by steps downstream in the pipeline")

}

// retrieve step metadata
func mtaBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "mtaBuild",
			Aliases:     []config.Alias{},
			Description: "Performs an mta build",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "mtarName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_mtarName"),
					},
					{
						Name:        "mtarGroup",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_mtarGroup"),
					},
					{
						Name: "version",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "artifactVersion"}},
						Default:   os.Getenv("PIPER_version"),
					},
					{
						Name:        "extensions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "extension"}},
						Default:     os.Getenv("PIPER_extensions"),
					},
					{
						Name:        "jobs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "jobs"}},
						Default:     0,
					},
					{
						Name:        "platform",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `CF`,
					},
					{
						Name:        "applicationName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationName"),
					},
					{
						Name:        "source",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `./`,
					},
					{
						Name:        "target",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `./`,
					},
					{
						Name:        "defaultNpmRegistry",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "npm/defaultNpmRegistry"}},
						Default:     os.Getenv("PIPER_defaultNpmRegistry"),
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/projectSettingsFile"}},
						Default:     os.Getenv("PIPER_projectSettingsFile"),
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/globalSettingsFile"}},
						Default:     os.Getenv("PIPER_globalSettingsFile"),
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
						Default:     os.Getenv("PIPER_m2Path"),
					},
					{
						Name:        "installArtifacts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "mtaDeploymentRepositoryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/mavenRepositoryPassword",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},

							{
								Name: "mtaDeploymentRepositoryPasswordId",
								Type: "secret",
							},

							{
								Name:    "mtaDeploymentRepositoryPasswordFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "mta-deployment-repository-passowrd",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_mtaDeploymentRepositoryPassword"),
					},
					{
						Name: "mtaDeploymentRepositoryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/mavenRepositoryUsername",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_mtaDeploymentRepositoryUser"),
					},
					{
						Name: "mtaDeploymentRepositoryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/mavenRepositoryURL",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_mtaDeploymentRepositoryUrl"),
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "mta/publish"}},
						Default:     false,
					},
					{
						Name:        "profiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name: "buildSettingsInfo",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/buildSettingsInfo",
							},
						},
						Scope:     []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_buildSettingsInfo"),
					},
					{
						Name:        "createBOM",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "enableSetTimestamp",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "createBuildArtifactsMetadata",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
				},
			},
			Containers: []config.Container{
				{Image: "devxci/mbtci-java11-node14"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "mtarFilePath"},
							{"name": "custom/mtaBuildToolDesc"},
							{"name": "custom/mtarPublishedUrl"},
							{"name": "custom/buildSettingsInfo"},
							{"name": "custom/mtaBuildArtifacts"},
						},
					},
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/TEST-*.xml", "type": "junit"},
							{"filePattern": "**/cobertura-coverage.xml", "type": "cobertura-coverage"},
							{"filePattern": "**/jacoco.xml", "type": "jacoco-coverage"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
