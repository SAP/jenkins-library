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
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type gradleExecuteBuildOptions struct {
	Path                          string   `json:"path,omitempty"`
	Task                          string   `json:"task,omitempty"`
	Publish                       bool     `json:"publish,omitempty"`
	RepositoryURL                 string   `json:"repositoryUrl,omitempty"`
	RepositoryPassword            string   `json:"repositoryPassword,omitempty"`
	RepositoryUsername            string   `json:"repositoryUsername,omitempty"`
	CreateBOM                     bool     `json:"createBOM,omitempty"`
	ArtifactVersion               string   `json:"artifactVersion,omitempty"`
	ArtifactGroupID               string   `json:"artifactGroupId,omitempty"`
	ArtifactID                    string   `json:"artifactId,omitempty"`
	UseWrapper                    bool     `json:"useWrapper,omitempty"`
	ApplyPublishingForAllProjects bool     `json:"applyPublishingForAllProjects,omitempty"`
	ExcludeCreateBOMForProjects   []string `json:"excludeCreateBOMForProjects,omitempty"`
	ExcludePublishingForProjects  []string `json:"excludePublishingForProjects,omitempty"`
	BuildFlags                    []string `json:"buildFlags,omitempty"`
}

type gradleExecuteBuildReports struct {
}

func (p *gradleExecuteBuildReports) persist(stepConfig gradleExecuteBuildOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/bom-gradle.xml", ParamRef: "", StepResultType: "sbom"},
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

type gradleExecuteBuildCommonPipelineEnvironment struct {
	custom struct {
		artifacts piperenv.Artifacts
	}
}

func (p *gradleExecuteBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "artifacts", value: p.custom.artifacts},
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

// GradleExecuteBuildCommand This step runs a gradle build command with parameters provided to the step.
func GradleExecuteBuildCommand() *cobra.Command {
	const STEP_NAME = "gradleExecuteBuild"

	metadata := gradleExecuteBuildMetadata()
	var stepConfig gradleExecuteBuildOptions
	var startTime time.Time
	var reports gradleExecuteBuildReports
	var commonPipelineEnvironment gradleExecuteBuildCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	var vaultClient config.VaultClient
	telemetryClient := &telemetry.Telemetry{}

	var createGradleExecuteBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step runs a gradle build command with parameters provided to the step.",
		Long:  `This step runs a gradle build command with parameters provided to the step.Supports execution of gradle tasks with or without wrapper.Gradle tasks and flags can be specified via 'task' or 'buildFlags' parameter. If both are not specified 'build' task will run by default.`,
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
			vaultClient = config.GlobalVaultClient()
			log.RegisterSecret(stepConfig.RepositoryPassword)
			log.RegisterSecret(stepConfig.RepositoryUsername)

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
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
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
						config.GlobalVaultClient(),
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
			defer vaultClient.MustRevokeToken()
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			gradleExecuteBuild(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGradleExecuteBuildFlags(createGradleExecuteBuildCmd, &stepConfig)
	return createGradleExecuteBuildCmd
}

func addGradleExecuteBuildFlags(cmd *cobra.Command, stepConfig *gradleExecuteBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.Path, "path", os.Getenv("PIPER_path"), "Path to the folder with build.gradle (or build.gradle.kts) file which should be executed.")
	cmd.Flags().StringVar(&stepConfig.Task, "task", `build`, "A single gradle task that should be executed. If you prefer more than one, use 'buildFlags' parameter. If 'buildFlags' parameter is specified, this parameter will be ignored.")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "Configures gradle to publish the artifact to a repository.")
	cmd.Flags().StringVar(&stepConfig.RepositoryURL, "repositoryUrl", os.Getenv("PIPER_repositoryUrl"), "Url to the repository to which the project artifacts should be published.")
	cmd.Flags().StringVar(&stepConfig.RepositoryPassword, "repositoryPassword", os.Getenv("PIPER_repositoryPassword"), "Password for the repository to which the project artifacts should be published.")
	cmd.Flags().StringVar(&stepConfig.RepositoryUsername, "repositoryUsername", os.Getenv("PIPER_repositoryUsername"), "Username for the repository to which the project artifacts should be published.")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Creates the bill of materials (BOM) using CycloneDX plugin.")
	cmd.Flags().StringVar(&stepConfig.ArtifactVersion, "artifactVersion", os.Getenv("PIPER_artifactVersion"), "Version of the artifact to be built.")
	cmd.Flags().StringVar(&stepConfig.ArtifactGroupID, "artifactGroupId", os.Getenv("PIPER_artifactGroupId"), "The group of the artifact.")
	cmd.Flags().StringVar(&stepConfig.ArtifactID, "artifactId", os.Getenv("PIPER_artifactId"), "The name of the artifact.")
	cmd.Flags().BoolVar(&stepConfig.UseWrapper, "useWrapper", false, "If set to false all commands are executed using 'gradle', otherwise 'gradlew' is executed.")
	cmd.Flags().BoolVar(&stepConfig.ApplyPublishingForAllProjects, "applyPublishingForAllProjects", false, "If set to false publishing logic will be applied in 'rootProject' directive, otherwise 'allprojects' will be directive used")
	cmd.Flags().StringSliceVar(&stepConfig.ExcludeCreateBOMForProjects, "excludeCreateBOMForProjects", []string{}, "Defines which projects/subprojects will be ignored during bom creation. Only if applyCreateBOMForAllProjects is set to true")
	cmd.Flags().StringSliceVar(&stepConfig.ExcludePublishingForProjects, "excludePublishingForProjects", []string{}, "Defines which projects/subprojects will be ignored during publishing. Only if applyCreateBOMForAllProjects is set to true")
	cmd.Flags().StringSliceVar(&stepConfig.BuildFlags, "buildFlags", []string{}, "Defines a list of tasks and/or arguments to be provided for gradle in the respective order to be executed. This list takes precedence if specified over 'task' parameter")

}

// retrieve step metadata
func gradleExecuteBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "gradleExecuteBuild",
			Aliases:     []config.Alias{},
			Description: "This step runs a gradle build command with parameters provided to the step.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "buildGradlePath"}},
						Default:     os.Getenv("PIPER_path"),
					},
					{
						Name:        "task",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `build`,
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "repositoryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_repositoryUrl"),
					},
					{
						Name: "repositoryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_repositoryPassword"),
					},
					{
						Name: "repositoryUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_repositoryUsername"),
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
						Name: "artifactVersion",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_artifactVersion"),
					},
					{
						Name: "artifactGroupId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "groupId",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_artifactGroupId"),
					},
					{
						Name: "artifactId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactId",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_artifactId"),
					},
					{
						Name:        "useWrapper",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "applyPublishingForAllProjects",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "excludeCreateBOMForProjects",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "excludePublishingForProjects",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "buildFlags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "gradle", Image: "gradle:6-jdk11-alpine"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/bom-gradle.xml", "type": "sbom"},
						},
					},
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "custom/artifacts", "type": "piperenv.Artifacts"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
