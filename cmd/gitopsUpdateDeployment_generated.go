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

type gitopsUpdateDeploymentOptions struct {
	BranchName            string   `json:"branchName,omitempty"`
	CommitMessage         string   `json:"commitMessage,omitempty"`
	ServerURL             string   `json:"serverUrl,omitempty"`
	Username              string   `json:"username,omitempty"`
	Password              string   `json:"password,omitempty"`
	FilePath              string   `json:"filePath,omitempty"`
	ContainerName         string   `json:"containerName,omitempty"`
	ContainerRegistryURL  string   `json:"containerRegistryUrl,omitempty"`
	ContainerImageNameTag string   `json:"containerImageNameTag,omitempty"`
	ChartPath             string   `json:"chartPath,omitempty"`
	HelmValues            []string `json:"helmValues,omitempty"`
	DeploymentName        string   `json:"deploymentName,omitempty"`
	Tool                  string   `json:"tool,omitempty"`
}

// GitopsUpdateDeploymentCommand Updates Kubernetes Deployment Manifest in an Infrastructure Git Repository
func GitopsUpdateDeploymentCommand() *cobra.Command {
	const STEP_NAME = "gitopsUpdateDeployment"

	metadata := gitopsUpdateDeploymentMetadata()
	var stepConfig gitopsUpdateDeploymentOptions
	var startTime time.Time

	var createGitopsUpdateDeploymentCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Updates Kubernetes Deployment Manifest in an Infrastructure Git Repository",
		Long: `This step allows you to update the deployment manifest for Kubernetes in a git repository.

It can for example be used for GitOps scenarios where the update of the manifests triggers an update of the corresponding deployment in Kubernetes.

As of today, it supports the update of deployment yaml files via kubectl patch and update a whole helm template.
For kubectl the container inside the yaml must be described within the following hierarchy: ` + "`" + `{"spec":{"template":{"spec":{"containers":[{...}]}}}}` + "`" + `
For helm the whole template is generated into a file and uploaded into the repository.`,
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
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			gitopsUpdateDeployment(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGitopsUpdateDeploymentFlags(createGitopsUpdateDeploymentCmd, &stepConfig)
	return createGitopsUpdateDeploymentCmd
}

func addGitopsUpdateDeploymentFlags(cmd *cobra.Command, stepConfig *gitopsUpdateDeploymentOptions) {
	cmd.Flags().StringVar(&stepConfig.BranchName, "branchName", `master`, "The name of the branch where the changes should get pushed into.")
	cmd.Flags().StringVar(&stepConfig.CommitMessage, "commitMessage", os.Getenv("PIPER_commitMessage"), "The commit message of the commit that will be done to do the changes.")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", `https://github.com`, "GitHub server url to the repository.")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User name for git authentication")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password/token for git authentication.")
	cmd.Flags().StringVar(&stepConfig.FilePath, "filePath", os.Getenv("PIPER_filePath"), "Relative path in the git repository to the deployment descriptor file that shall be updated")
	cmd.Flags().StringVar(&stepConfig.ContainerName, "containerName", os.Getenv("PIPER_containerName"), "The name of the container to update")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryURL, "containerRegistryUrl", os.Getenv("PIPER_containerRegistryUrl"), "http(s) url of the Container registry where the image is located")
	cmd.Flags().StringVar(&stepConfig.ContainerImageNameTag, "containerImageNameTag", os.Getenv("PIPER_containerImageNameTag"), "Container image name with version tag to annotate in the deployment configuration.")
	cmd.Flags().StringVar(&stepConfig.ChartPath, "chartPath", os.Getenv("PIPER_chartPath"), "Defines the chart path for deployments using helm.")
	cmd.Flags().StringSliceVar(&stepConfig.HelmValues, "helmValues", []string{}, "List of helm values as YAML file reference or URL (as per helm parameter description for `-f` / `--values`)")
	cmd.Flags().StringVar(&stepConfig.DeploymentName, "deploymentName", os.Getenv("PIPER_deploymentName"), "Defines the name of the deployment.")
	cmd.Flags().StringVar(&stepConfig.Tool, "tool", `kubectl`, "Defines the tool which should be used to update the deployment description.")

	cmd.MarkFlagRequired("branchName")
	cmd.MarkFlagRequired("serverUrl")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("filePath")
	cmd.MarkFlagRequired("containerRegistryUrl")
	cmd.MarkFlagRequired("containerImageNameTag")
	cmd.MarkFlagRequired("tool")
}

// retrieve step metadata
func gitopsUpdateDeploymentMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "gitopsUpdateDeployment",
			Aliases:     []config.Alias{},
			Description: "Updates Kubernetes Deployment Manifest in an Infrastructure Git Repository",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "gitHttpsCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing username/password for http access to your git repository.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "branchName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     `master`,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "commitMessage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_commitMessage"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     `https://github.com`,
						Aliases:     []config.Alias{{Name: "githubServerUrl"}},
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "gitHttpsCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Default:   os.Getenv("PIPER_username"),
						Aliases:   []config.Alias{},
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "gitHttpsCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Default:   os.Getenv("PIPER_password"),
						Aliases:   []config.Alias{},
					},
					{
						Name:        "filePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     os.Getenv("PIPER_filePath"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "containerName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_containerName"),
						Aliases:     []config.Alias{},
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
						Mandatory: true,
						Default:   os.Getenv("PIPER_containerRegistryUrl"),
						Aliases:   []config.Alias{{Name: "dockerRegistryUrl"}},
					},
					{
						Name: "containerImageNameTag",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/imageNameTag",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Default:   os.Getenv("PIPER_containerImageNameTag"),
						Aliases:   []config.Alias{{Name: "image"}, {Name: "containerImage"}},
					},
					{
						Name:        "chartPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_chartPath"),
						Aliases:     []config.Alias{{Name: "helmChartPath"}},
					},
					{
						Name:        "helmValues",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Default:     []string{},
						Aliases:     []config.Alias{},
					},
					{
						Name:        "deploymentName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_deploymentName"),
						Aliases:     []config.Alias{{Name: "helmDeploymentName"}},
					},
					{
						Name:        "tool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Default:     `kubectl`,
						Aliases:     []config.Alias{},
					},
				},
			},
			Containers: []config.Container{
				{Image: "dtzar/helm-kubectl:3.3.4", WorkingDir: "/config", Options: []config.Option{{Name: "-u", Value: "0"}}, Conditions: []config.Condition{{ConditionRef: "strings-equal", Params: []config.Param{{Name: "tool", Value: "helm"}}}}},
				{Image: "dtzar/helm-kubectl:2.12.1", WorkingDir: "/config", Options: []config.Option{{Name: "-u", Value: "0"}}, Conditions: []config.Condition{{ConditionRef: "strings-equal", Params: []config.Param{{Name: "tool", Value: "kubectl"}}}}},
			},
		},
	}
	return theMetaData
}
