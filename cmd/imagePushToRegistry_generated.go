// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type imagePushToRegistryOptions struct {
	TargetImages           []map[string]interface{} `json:"targetImages,omitempty"`
	SourceImages           []string                 `json:"sourceImages,omitempty"`
	SourceRegistryURL      string                   `json:"sourceRegistryUrl,omitempty"`
	SourceRegistryUser     string                   `json:"sourceRegistryUser,omitempty"`
	SourceRegistryPassword string                   `json:"sourceRegistryPassword,omitempty"`
	TargetRegistryURL      string                   `json:"targetRegistryUrl,omitempty"`
	TargetRegistryUser     string                   `json:"targetRegistryUser,omitempty"`
	TargetRegistryPassword string                   `json:"targetRegistryPassword,omitempty"`
	TagLatest              bool                     `json:"tagLatest,omitempty"`
	TagArtifactVersion     bool                     `json:"tagArtifactVersion,omitempty"`
	DockerConfigJSON       string                   `json:"dockerConfigJSON,omitempty"`
	LocalDockerImagePath   string                   `json:"localDockerImagePath,omitempty"`
	TargetArchitecture     string                   `json:"targetArchitecture,omitempty"`
	ImageTag               string                   `json:"imageTag,omitempty"`
}

// ImagePushToRegistryCommand Allows you to copy a Docker image from a source container registry  to a destination container registry.
func ImagePushToRegistryCommand() *cobra.Command {
	const STEP_NAME = "imagePushToRegistry"

	metadata := imagePushToRegistryMetadata()
	var stepConfig imagePushToRegistryOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createImagePushToRegistryCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Allows you to copy a Docker image from a source container registry  to a destination container registry.",
		Long: `In case you want to pull an existing image from a remote container registry, a source image and source registry needs to be specified.<br />
This makes it possible to move an image from one registry to another.

The imagePushToRegistry is not similar in functionality to containerPushToRegistry (which is currently a groovy based step and only be used in jenkins).
Currently the imagePushToRegistry only supports copying a local image or image from source remote registry to destination registry.`,
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
			log.RegisterSecret(stepConfig.SourceRegistryUser)
			log.RegisterSecret(stepConfig.SourceRegistryPassword)
			log.RegisterSecret(stepConfig.TargetRegistryUser)
			log.RegisterSecret(stepConfig.TargetRegistryPassword)
			log.RegisterSecret(stepConfig.DockerConfigJSON)

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
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			imagePushToRegistry(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addImagePushToRegistryFlags(createImagePushToRegistryCmd, &stepConfig)
	return createImagePushToRegistryCmd
}

func addImagePushToRegistryFlags(cmd *cobra.Command, stepConfig *imagePushToRegistryOptions) {

	cmd.Flags().StringSliceVar(&stepConfig.SourceImages, "sourceImages", []string{}, "Defines the names (incl. tag) of the images that will be pulled from source registry. This is helpful for moving images from one location to another.\nPlease ensure that targetImages and sourceImages correspond to each other: the first image in sourceImages will be mapped to the first image in the targetImages parameter.\n")
	cmd.Flags().StringVar(&stepConfig.SourceRegistryURL, "sourceRegistryUrl", os.Getenv("PIPER_sourceRegistryUrl"), "Defines a registry url from where the image should optionally be pulled from, incl. the protocol like `https://my.registry.com`*\"")
	cmd.Flags().StringVar(&stepConfig.SourceRegistryUser, "sourceRegistryUser", os.Getenv("PIPER_sourceRegistryUser"), "Username of the source registry where the image should be pushed pulled from.")
	cmd.Flags().StringVar(&stepConfig.SourceRegistryPassword, "sourceRegistryPassword", os.Getenv("PIPER_sourceRegistryPassword"), "Password of the source registry where the image should be pushed pulled from.")
	cmd.Flags().StringVar(&stepConfig.TargetRegistryURL, "targetRegistryUrl", os.Getenv("PIPER_targetRegistryUrl"), "Defines a registry url from where the image should optionally be pushed to, incl. the protocol like `https://my.registry.com`*\"")
	cmd.Flags().StringVar(&stepConfig.TargetRegistryUser, "targetRegistryUser", os.Getenv("PIPER_targetRegistryUser"), "Username of the target registry where the image should be pushed to.")
	cmd.Flags().StringVar(&stepConfig.TargetRegistryPassword, "targetRegistryPassword", os.Getenv("PIPER_targetRegistryPassword"), "Password of the target registry where the image should be pushed to.")
	cmd.Flags().BoolVar(&stepConfig.TagLatest, "tagLatest", false, "Defines if the image should be tagged as `latest`")
	cmd.Flags().BoolVar(&stepConfig.TagArtifactVersion, "tagArtifactVersion", false, "The parameter is not supported yet. Defines if the image should be tagged with the artifact version")
	cmd.Flags().StringVar(&stepConfig.DockerConfigJSON, "dockerConfigJSON", os.Getenv("PIPER_dockerConfigJSON"), "Path to the file `.docker/config.json` - this is typically provided by your CI/CD system. You can find more details about the Docker credentials in the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/).")
	cmd.Flags().StringVar(&stepConfig.LocalDockerImagePath, "localDockerImagePath", os.Getenv("PIPER_localDockerImagePath"), "If the `localDockerImagePath` is a directory, it will be read as an OCI image layout. Otherwise, `localDockerImagePath` is assumed to be a docker-style tarball.")
	cmd.Flags().StringVar(&stepConfig.TargetArchitecture, "targetArchitecture", os.Getenv("PIPER_targetArchitecture"), "Specifies the targetArchitecture in the form os/arch[/variant][:osversion] (e.g. linux/amd64). All OS and architectures of the specified image will be copied if it is a multi-platform image. To only push a single platform to the target registry use this parameter")
	cmd.Flags().StringVar(&stepConfig.ImageTag, "imageTag", os.Getenv("PIPER_imageTag"), "Tag of the image which will be built")

	cmd.MarkFlagRequired("sourceImages")
	cmd.MarkFlagRequired("sourceRegistryUrl")
	cmd.MarkFlagRequired("targetRegistryUrl")
	cmd.MarkFlagRequired("targetRegistryUser")
	cmd.MarkFlagRequired("targetRegistryPassword")
}

// retrieve step metadata
func imagePushToRegistryMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "imagePushToRegistry",
			Aliases:     []config.Alias{},
			Description: "Allows you to copy a Docker image from a source container registry  to a destination container registry.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Resources: []config.StepResources{
					{Name: "source", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "targetImages",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name: "sourceImages",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/imageNameTags",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "[]string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   []string{},
					},
					{
						Name: "sourceRegistryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/registryUrl",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_sourceRegistryUrl"),
					},
					{
						Name: "sourceRegistryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/repositoryUsername",
							},

							{
								Name:    "registryCredentialsVaultSecretName",
								Type:    "vaultSecret",
								Default: "docker-registry",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_sourceRegistryUser"),
					},
					{
						Name: "sourceRegistryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/repositoryPassword",
							},

							{
								Name:    "registryCredentialsVaultSecretName",
								Type:    "vaultSecret",
								Default: "docker-registry",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_sourceRegistryPassword"),
					},
					{
						Name:        "targetRegistryUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_targetRegistryUrl"),
					},
					{
						Name: "targetRegistryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "registryCredentialsVaultSecretName",
								Type:    "vaultSecret",
								Default: "docker-registry",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRegistryUser"),
					},
					{
						Name: "targetRegistryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "registryCredentialsVaultSecretName",
								Type:    "vaultSecret",
								Default: "docker-registry",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRegistryPassword"),
					},
					{
						Name:        "tagLatest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "tagArtifactVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "dockerConfigJSON",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "dockerConfigFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "docker-config",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_dockerConfigJSON"),
					},
					{
						Name:        "localDockerImagePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_localDockerImagePath"),
					},
					{
						Name:        "targetArchitecture",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_targetArchitecture"),
					},
					{
						Name: "imageTag",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "artifactVersion"}, {Name: "containerImageTag"}},
						Default:   os.Getenv("PIPER_imageTag"),
					},
				},
			},
			Containers: []config.Container{
				{Image: "gcr.io/go-containerregistry/crane:debug", EnvVars: []config.EnvVar{{Name: "container", Value: "docker"}}, Options: []config.Option{{Name: "-u", Value: "0"}, {Name: "--entrypoint", Value: ""}}},
			},
		},
	}
	return theMetaData
}
