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

type githubSetCommitStatusOptions struct {
	APIURL      string `json:"apiUrl,omitempty"`
	CommitID    string `json:"commitId,omitempty"`
	Context     string `json:"context,omitempty"`
	Description string `json:"description,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Status      string `json:"status,omitempty"`
	TargetURL   string `json:"targetUrl,omitempty"`
	Token       string `json:"token,omitempty"`
}

// GithubSetCommitStatusCommand Set a status of a certain commit.
func GithubSetCommitStatusCommand() *cobra.Command {
	const STEP_NAME = "githubSetCommitStatus"

	metadata := githubSetCommitStatusMetadata()
	var stepConfig githubSetCommitStatusOptions
	var startTime time.Time

	var createGithubSetCommitStatusCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Set a status of a certain commit.",
		Long: `This step allows you to set a status for a certain commit.
Details can be found here: https://developer.github.com/v3/repos/statuses/.

Typically, following information is set:

* state (pending, failure, success)
* context
* target url (linkt to details)

It can for example be used to create additional check indicators for a pull request which can be evaluated and also be enforced by GitHub configuration.`,
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
			log.RegisterSecret(stepConfig.Token)

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
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			githubSetCommitStatus(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGithubSetCommitStatusFlags(createGithubSetCommitStatusCmd, &stepConfig)
	return createGithubSetCommitStatusCmd
}

func addGithubSetCommitStatusFlags(cmd *cobra.Command, stepConfig *githubSetCommitStatusOptions) {
	cmd.Flags().StringVar(&stepConfig.APIURL, "apiUrl", `https://api.github.com`, "Set the GitHub API url.")
	cmd.Flags().StringVar(&stepConfig.CommitID, "commitId", os.Getenv("PIPER_commitId"), "The commitId for which the status should be set.")
	cmd.Flags().StringVar(&stepConfig.Context, "context", os.Getenv("PIPER_context"), "Label for the status which will for example show up in a pull request.")
	cmd.Flags().StringVar(&stepConfig.Description, "description", os.Getenv("PIPER_description"), "Short description of the status.")
	cmd.Flags().StringVar(&stepConfig.Owner, "owner", os.Getenv("PIPER_owner"), "Name of the GitHub organization.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Name of the GitHub repository.")
	cmd.Flags().StringVar(&stepConfig.Status, "status", os.Getenv("PIPER_status"), "Status which should be set on the commitId.")
	cmd.Flags().StringVar(&stepConfig.TargetURL, "targetUrl", os.Getenv("PIPER_targetUrl"), "Target url to associate the status with.")
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line.")

	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("commitId")
	cmd.MarkFlagRequired("context")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("status")
	cmd.MarkFlagRequired("token")
}

// retrieve step metadata
func githubSetCommitStatusMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "githubSetCommitStatus",
			Aliases: []config.Alias{},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "apiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubApiUrl"}},
					},
					{
						Name:        "commitId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "context",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "description",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "owner",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubOrg"}},
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubRepo"}},
					},
					{
						Name:        "status",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "targetUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "token",
						ResourceRef: []config.ResourceReference{{Name: "githubTokenCredentialsId", Param: ""}},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubToken"}},
					},
				},
			},
		},
	}
	return theMetaData
}
