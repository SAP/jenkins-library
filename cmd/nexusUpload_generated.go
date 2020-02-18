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

type nexusUploadOptions struct {
	NexusVersion int    `json:"nexusVersion,omitempty"`
	Url          string `json:"url,omitempty"`
	GroupID      string `json:"groupId,omitempty"`
	Version      string `json:"version,omitempty"`
	Repository   string `json:"repository,omitempty"`
	Artifacts    string `json:"artifacts,omitempty"`
	User         string `json:"user,omitempty"`
	Password     string `json:"password,omitempty"`
}

// NexusUploadCommand Upload to Nexus RM
func NexusUploadCommand() *cobra.Command {
	metadata := nexusUploadMetadata()
	var stepConfig nexusUploadOptions
	var startTime time.Time

	var createNexusUploadCmd = &cobra.Command{
		Use:   "nexusUpload",
		Short: "Upload to Nexus RM",
		Long:  `Upload artifacts to Nexus Repository Manager`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("nexusUpload")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "nexusUpload", &stepConfig, config.OpenPiperFile)
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
			telemetry.Initialize(GeneralConfig.NoTelemetry, "nexusUpload")
			nexusUpload(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addNexusUploadFlags(createNexusUploadCmd, &stepConfig)
	return createNexusUploadCmd
}

func addNexusUploadFlags(cmd *cobra.Command, stepConfig *nexusUploadOptions) {
	cmd.Flags().IntVar(&stepConfig.NexusVersion, "nexusVersion", 3, "The nexus Repository Manager version.")
	cmd.Flags().StringVar(&stepConfig.Url, "url", os.Getenv("PIPER_url"), "URL of the nexus. The scheme part of the URL will not be considered, because only http is supported.")
	cmd.Flags().StringVar(&stepConfig.GroupID, "groupId", os.Getenv("PIPER_groupId"), "Group ID of the artifacts.")
	cmd.Flags().StringVar(&stepConfig.Version, "version", os.Getenv("PIPER_version"), "Version of the artifacts.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Name of the nexus repository.")
	cmd.Flags().StringVar(&stepConfig.Artifacts, "artifacts", os.Getenv("PIPER_artifacts"), "JSON encoded list of artifact descriptions.")
	cmd.Flags().StringVar(&stepConfig.User, "user", os.Getenv("PIPER_user"), "User")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password")

	cmd.MarkFlagRequired("url")
	cmd.MarkFlagRequired("groupId")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("artifacts")
}

// retrieve step metadata
func nexusUploadMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "nexusVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "url",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "groupId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "version",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "artifacts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "user",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
