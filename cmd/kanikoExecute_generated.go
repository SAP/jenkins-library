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

type kanikoExecuteOptions struct {
	BuildOptions                     []string `json:"buildOptions,omitempty"`
	BuildSettingsInfo                string   `json:"buildSettingsInfo,omitempty"`
	ContainerMultiImageBuild         bool     `json:"containerMultiImageBuild,omitempty"`
	ContainerMultiImageBuildExcludes []string `json:"containerMultiImageBuildExcludes,omitempty"`
	ContainerBuildOptions            string   `json:"containerBuildOptions,omitempty"`
	ContainerImage                   string   `json:"containerImage,omitempty"`
	ContainerImageName               string   `json:"containerImageName,omitempty"`
	ContainerImageTag                string   `json:"containerImageTag,omitempty"`
	ContainerPreparationCommand      string   `json:"containerPreparationCommand,omitempty"`
	ContainerRegistryURL             string   `json:"containerRegistryUrl,omitempty"`
	CustomTLSCertificateLinks        []string `json:"customTlsCertificateLinks,omitempty"`
	DockerConfigJSON                 string   `json:"dockerConfigJSON,omitempty"`
	DockerfilePath                   string   `json:"dockerfilePath,omitempty"`
	TargetArchitectures              []string `json:"targetArchitectures,omitempty"`
}

type kanikoExecuteCommonPipelineEnvironment struct {
	container struct {
		registryURL   string
		imageNameTag  string
		imageNames    []string
		imageNameTags []string
	}
	custom struct {
		buildSettingsInfo string
	}
}

func (p *kanikoExecuteCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "container", name: "registryUrl", value: p.container.registryURL},
		{category: "container", name: "imageNameTag", value: p.container.imageNameTag},
		{category: "container", name: "imageNames", value: p.container.imageNames},
		{category: "container", name: "imageNameTags", value: p.container.imageNameTags},
		{category: "custom", name: "buildSettingsInfo", value: p.custom.buildSettingsInfo},
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

// KanikoExecuteCommand Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.
func KanikoExecuteCommand() *cobra.Command {
	const STEP_NAME = "kanikoExecute"

	metadata := kanikoExecuteMetadata()
	var stepConfig kanikoExecuteOptions
	var startTime time.Time
	var commonPipelineEnvironment kanikoExecuteCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createKanikoExecuteCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.",
		Long: `Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.

### Building multiple container images

The step allows you to build multiple container images with one run.
This is suitable in case you need to create multiple images for one microservice, e.g. for testing.

All images will get the same "root" name and the same versioning.<br />
**Thus, this is not suitable to be used for a monorepo approach!** For monorepos you need to use a build tool natively capable to take care for monorepos
or implement a custom logic and for example execute this ` + "`" + `kanikoExecute` + "`" + ` step multiple times in your custom pipeline.

You can activate multiple builds using the parameters

* [containerMultiImageBuild](#containermultiimagebuild) for activation
* [containerMultiImageBuildExcludes](#containermultiimagebuildexcludes) for defining excludes`,
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
			log.RegisterSecret(stepConfig.DockerConfigJSON)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient = &splunk.Splunk{}
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
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			kanikoExecute(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addKanikoExecuteFlags(createKanikoExecuteCmd, &stepConfig)
	return createKanikoExecuteCmd
}

func addKanikoExecuteFlags(cmd *cobra.Command, stepConfig *kanikoExecuteOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.BuildOptions, "buildOptions", []string{`--skip-tls-verify-pull`, `--ignore-path=/`}, "Defines a list of build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.")
	cmd.Flags().StringVar(&stepConfig.BuildSettingsInfo, "buildSettingsInfo", os.Getenv("PIPER_buildSettingsInfo"), "Build settings info is typically filled by the step automatically to create information about the build settings that were used during the mta build. This information is typically used for compliance related processes.")
	cmd.Flags().BoolVar(&stepConfig.ContainerMultiImageBuild, "containerMultiImageBuild", false, "Defines if multiple containers should be build. Dockerfiles are used using the pattern **/Dockerfile*. Excludes can be defined via [`containerMultiImageBuildExcludes`](#containermultiimagebuildexscludes).")
	cmd.Flags().StringSliceVar(&stepConfig.ContainerMultiImageBuildExcludes, "containerMultiImageBuildExcludes", []string{}, "Defines a list of Dockerfile paths to exclude from the build when using [`containerMultiImageBuild`](#containermultiimagebuild).")
	cmd.Flags().StringVar(&stepConfig.ContainerBuildOptions, "containerBuildOptions", os.Getenv("PIPER_containerBuildOptions"), "Deprected, please use buildOptions. Defines the build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.")
	cmd.Flags().StringVar(&stepConfig.ContainerImage, "containerImage", os.Getenv("PIPER_containerImage"), "Defines the full name of the Docker image to be created including registry, image name and tag like `my.docker.registry/path/myImageName:myTag`. If left empty, image will not be pushed.")
	cmd.Flags().StringVar(&stepConfig.ContainerImageName, "containerImageName", os.Getenv("PIPER_containerImageName"), "Name of the container which will be built - will be used instead of parameter `containerImage`")
	cmd.Flags().StringVar(&stepConfig.ContainerImageTag, "containerImageTag", os.Getenv("PIPER_containerImageTag"), "Tag of the container which will be built - will be used instead of parameter `containerImage`")
	cmd.Flags().StringVar(&stepConfig.ContainerPreparationCommand, "containerPreparationCommand", `rm -f /kaniko/.docker/config.json`, "Defines the command to prepare the Kaniko container. By default the contained credentials are removed in order to allow anonymous access to container registries.")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryURL, "containerRegistryUrl", os.Getenv("PIPER_containerRegistryUrl"), "http(s) url of the Container registry where the image should be pushed to - will be used instead of parameter `containerImage`")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List containing download links of custom TLS certificates. This is required to ensure trusted connections to registries with custom certificates.")
	cmd.Flags().StringVar(&stepConfig.DockerConfigJSON, "dockerConfigJSON", os.Getenv("PIPER_dockerConfigJSON"), "Path to the file `.docker/config.json` - this is typically provided by your CI/CD system. You can find more details about the Docker credentials in the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/).")
	cmd.Flags().StringVar(&stepConfig.DockerfilePath, "dockerfilePath", `Dockerfile`, "Defines the location of the Dockerfile relative to the Jenkins workspace.")
	cmd.Flags().StringSliceVar(&stepConfig.TargetArchitectures, "targetArchitectures", []string{``}, "Defines the target architectures for which the build should run using OS and architecture separated by a comma. (EXPERIMENTAL)")

}

// retrieve step metadata
func kanikoExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "kanikoExecute",
			Aliases:     []config.Alias{},
			Description: "Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "dockerConfigJsonCredentialsId", Description: "Jenkins 'Secret file' credentials ID containing Docker config.json (with registry credential(s)). You can create it like explained in the [protocodeExecuteScan Prerequisites section](https://www.project-piper.io/steps/protecodeExecuteScan/#prerequisites).", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "buildOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{`--skip-tls-verify-pull`, `--ignore-path=/`},
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
						Name:        "containerMultiImageBuild",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "containerMultiImageBuildExcludes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "containerBuildOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_containerBuildOptions"),
					},
					{
						Name:        "containerImage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "containerImageNameAndTag", Deprecated: true}},
						Default:     os.Getenv("PIPER_containerImage"),
					},
					{
						Name:        "containerImageName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerImageName"}},
						Default:     os.Getenv("PIPER_containerImageName"),
					},
					{
						Name: "containerImageTag",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "artifactVersion"}},
						Default:   os.Getenv("PIPER_containerImageTag"),
					},
					{
						Name:        "containerPreparationCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `rm -f /kaniko/.docker/config.json`,
					},
					{
						Name: "containerRegistryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/registryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "dockerRegistryUrl"}},
						Default:   os.Getenv("PIPER_containerRegistryUrl"),
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name: "dockerConfigJSON",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/dockerConfigJSON",
							},

							{
								Name: "dockerConfigJsonCredentialsId",
								Type: "secret",
							},

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
						Name:        "dockerfilePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerfile"}},
						Default:     `Dockerfile`,
					},
					{
						Name:        "targetArchitectures",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{``},
					},
				},
			},
			Containers: []config.Container{
				{Image: "gcr.io/kaniko-project/executor:debug", EnvVars: []config.EnvVar{{Name: "container", Value: "docker"}, {Name: "TMPDIR", Value: "/"}}, Options: []config.Option{{Name: "-u", Value: "0"}, {Name: "--entrypoint", Value: ""}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "container/registryUrl"},
							{"name": "container/imageNameTag"},
							{"name": "container/imageNames", "type": "[]string"},
							{"name": "container/imageNameTags", "type": "[]string"},
							{"name": "custom/buildSettingsInfo"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
